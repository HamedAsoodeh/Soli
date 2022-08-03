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
	txShares, evdShares, msgLens := rawShareCount(data, txConf)

	msgShares := 0
	for _, msgLen := range msgLens {
		msgShares += msgLen
	}

	// calculate the smallest possible square size
	squareSize := int(types.NextHighestPowerOf2(uint64(txShares + evdShares + msgShares)))

	// incrememtally check square sizes by adding padding to messages in order
	// to account for the non-interactive default rules
	for fits := false; !fits; {
		fits = checkFitInSquare(txShares+evdShares, squareSize, msgLens...)
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
func rawShareCount(data *core.Data, txConf client.TxConfig) (txShares, evdShares int, msgLens []int) {
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

	return txShares, evdShares, msgLens
}

// checkFitInSquare uses the non interactive default rules to see if messages of
// some lengths will fit in a square of size origSquareSize starting at share
// index cursor. See non-interactive default rules
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#non-interactive-default-rules
func checkFitInSquare(cursor, origSquareSize int, msgLens ...int) (fits bool) {
	for _, msgLen := range msgLens {
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

// nextAlignedPowerOfTwo calculates the next index in a row that is an aligned
// power of two or returns false is the msg cannot fit on the given row at the
// next aligned power of two. An aligned power of two means that the largest
// power of two that fits entirely in the msg or the square size. pls see specs
// for further details. Assumes that cursor < k, all args are non negative, and
// that k is a power of two.
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#non-interactive-default-rules
func nextAlignedPowerOfTwo(cursor, msgLen, k int) (int, bool) {
	// if we're starting at the beginning of the row, then return as there are
	// no cases where we don't. This check is redundant to one performed in
	// checkFitsInRow, but just to explicit and future proof, this has been left
	// in.
	if cursor == 0 {
		return cursor, true
	}
	// if the aligned power of two is larger than the room left in the row, then
	// the msg will not fit. We add 1 here to adjust for cursor being 0 indexed.
	nextLowest := nextLowestPowerOfTwo(msgLen)
	if k-nextLowest < cursor {
		return 0, false
	}
	// round up to nearest aligned power of two
	cursor = roundUpBy(cursor, nextLowest)
	if cursor+msgLen > k {
		return 0, false
	}
	return cursor, true
}

// roundUpBy rounds cursor up to the next interval of v. If cursor is divisible
// by v, then it returns cursor
func roundUpBy(cursor, v int) int {
	switch {
	case cursor == 0:
		return cursor
	case cursor%v == 0:
		return cursor
	default:
		return ((cursor / v) + 1) * v
	}
}

func nextPowerOfTwo(v int) int {
	k := 1
	for k < v {
		k = k << 1
	}
	return k
}

func nextLowestPowerOfTwo(v int) int {
	c := nextPowerOfTwo(v)
	if c == v {
		return c
	}
	return c / 2
}
