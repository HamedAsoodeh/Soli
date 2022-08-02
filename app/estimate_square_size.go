package app

import (
	"bytes"
	"sort"

	"github.com/celestiaorg/celestia-app/pkg/inclusion/appconsts"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/tendermint/tendermint/pkg/consts"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
)

// estimateSquareSize uses the provided block data to estimate the square size
// assuming that all malleated txs follow the non interactive default rules.
func estimateSquareSize(data *core.Data, txConf client.TxConfig) uint64 {
	// get the raw count of shares taken by each type of block data
	txShares, evdShares, msgShareCounts := rawShareCount(data, txConf)

	msgShares := 0
	for _, msgLen := range msgShareCounts {
		msgShares += msgLen
	}

	// calculate the smallest possible square size
	squareSize := int(types.NextHighestPowerOf2(uint64(txShares + evdShares + msgShares)))

	// incrememtally check square sizes by adding padding to messages in order
	// to account for the non-interactive default rules
	for fits := false; !fits; {
		fits = estimateNonInteractiveDefaultPadding(txShares+evdShares, squareSize, msgShareCounts)
		// increment the square size
		squareSize = int(types.NextHighestPowerOf2(uint64(squareSize) + 1))
		if squareSize >= consts.MaxSquareSize {
			return consts.MaxSquareSize
		}
	}

	switch {
	case squareSize < consts.MinSquareSize:
		return consts.MinSquareSize
	default:
		return uint64(squareSize)
	}
}

// rawShareCount calculates the number of shares taken by all of the included
// txs, evidence, and each msg.
//
// NOTE: It assumes that every malleatable tx has a viable commit for whatever
// square size that we end up picking. This is a flaw in this estimation
// algorithm, where someone can submit a max sized wPFD to the mempool, and as
// long as there are other wPFDs in the mempool, will force all block producers
// using this code to produce a block with the max square size.
func rawShareCount(data *core.Data, txConf client.TxConfig) (txShares, evdShares int, msgShareCounts []int) {
	// msgSummary is used to keep track fo the size and the namespace so that we
	// can sort the namespaces before returning
	type msgSummary struct {
		size      int
		namespace []byte
	}

	var msgSummaries []msgSummary

	// we use bytes instead of shares for tx and evd as they are encoded
	// contiguously in the square, unlike msgs where each of which is assigned their
	// own set of shares
	txBytes, evdBytes := 0, 0
	for _, rawTx := range data.Txs {
		// decode the Tx
		tx, err := txConf.TxDecoder()(rawTx)
		if err != nil {
			continue
		}

		authTx, ok := tx.(signing.Tx)
		if !ok {
			continue
		}

		wireMsg, err := types.ExtractMsgWirePayForData(authTx)
		if err != nil {
			// we catch this error because it means that there are no
			// potentially valid MsgWirePayForData messages in this tx. If the
			// tx doesn't have a wirePFD, then it won't contribute any message
			// shares to the block, and since we're only estimating here, we can
			// move on without handling or bubbling the error.
			txBytes += len(rawTx)
			continue
		}

		// if the there is a malleated txs, then we should also account for the
		// wrapped tx bytes
		txBytes += len(rawTx) + appconsts.MalleatedTxBytes

		msgSummaries = append(msgSummaries, msgSummary{types.MsgSharesUsed(int(wireMsg.MessageSize)), wireMsg.MessageNameSpaceId})
	}

	txShares = txBytes / consts.TxShareSize
	if txBytes > 0 {
		txShares++ // add one to round up
	}

	for _, evd := range data.Evidence.Evidence {
		evdBytes += evd.Size() + types.DelimLen(uint64(evd.Size()))
	}
	evdShares = evdBytes / consts.TxShareSize
	if evdBytes > 0 {
		evdShares++ // add one to round up
	}

	// sort the msgSummaries in order to order properly
	sort.Slice(msgSummaries, func(i, j int) bool {
		return bytes.Compare(msgSummaries[i].namespace, msgSummaries[j].namespace) < 0
	})

	// isolate the sizes as we no longer need the namespaces
	msgShares := make([]int, len(msgSummaries))
	for i, summary := range msgSummaries {
		msgShares[i] = summary.size
	}

	return txShares, evdShares, msgShareCounts
}

// estimateNonInteractiveDefaultPadding uses the non interactive default rules
// to estimate the total number of shares required by a set of messages.
func estimateNonInteractiveDefaultPadding(cursor, origSquareSize int, msgShareCounts []int) (fits bool) {
	for _, msgLen := range msgShareCounts {
		currentRow := (cursor / origSquareSize)
		currentCol := cursor % origSquareSize
		switch {
		// check if we're finished
		case currentRow >= origSquareSize:
			return false
		// we overflow to the next row, so start at the next row
		case (currentCol + msgLen) > origSquareSize:
			cursor = (origSquareSize * (currentRow + 1)) - 1 + msgLen
		// the msg fits on this row, therefore increase the cursor by msgLen
		default:
			cursor += msgLen
		}
	}
	// this check catches the edge case where the last message overflows rows
	if cursor/origSquareSize >= origSquareSize {
		return false
	}
	return true
}
