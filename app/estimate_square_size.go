package app

import (
	"bytes"
	"sort"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/tendermint/tendermint/pkg/consts"
	coretypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

// estimateSquareSize uses the provided block data to estimate the square size
// assuming that all malleated txs follow the non interactive default rules.
func estimateSquareSize(txs []parsedTx, evd coretypes.EvidenceList) uint64 {
	// get the raw count of shares taken by each type of block data
	txShares, evdShares, msgLens := rawShareCount(txs, evd)
	msgShares := 0
	for _, msgLen := range msgLens {
		msgShares += msgLen
	}

	// calculate the smallest possible square size that could contian all the
	// messages
	squareSize := nextPowerOfTwo(txShares + evdShares + msgShares)
	// the starting square size should be the minimum
	if squareSize < consts.MinSquareSize {
		squareSize = int(consts.MinSquareSize)
	}

	for {
		// assume that all the msgs in the square use the non-interactive
		// default rules and see if we can fit them in the smallest starting
		// square size. We start the cusor (share index) at the begginning of
		// the message shares (txShares+evdShares), because shares that do not
		// follow the non-interactive defaults are simple to estimate.
		fits := shares.FitsInSquare(txShares+evdShares, squareSize, msgLens...)
		switch {
		// stop estimating if we know we can reach the max square size
		case squareSize >= consts.MaxSquareSize:
			return consts.MaxSquareSize
		// return if we've found a square size that fits all of the txs
		case fits:
			return uint64(squareSize)
		// try the next largest square size if we can't fit all the txs
		case !fits:
			// increment the square size
			squareSize = int(nextPowerOfTwo(squareSize + 1))
		}
	}
}

// rawShareCount calculates the number of shares taken by all of the included
// txs, evidence, and each msg.
func rawShareCount(txs []parsedTx, evd coretypes.EvidenceList) (txShares, evdShares int, msgLens []int) {
	// msgSummary is used to keep track fo the size and the namespace so that we
	// can sort the namespaces before returning.
	type msgSummary struct {
		size      int
		namespace []byte
	}

	var msgSummaries []msgSummary

	// we use bytes instead of shares for tx and evd as they are encoded
	// contiguously in the square, unlike msgs where each of which is assigned their
	// own set of shares
	txBytes, evdBytes := 0, 0
	for _, pTx := range txs {
		// if there is no wire message in this tx, then we can simply add the
		// bytes and move on.
		if pTx.msg == nil {
			txBytes += len(pTx.rawTx)
			continue
		}

		// if the there is a malleated txs, then we should also account for the
		// wrapped tx bytes
		txBytes += appconsts.MalleatedTxBytes

		msgSummaries = append(msgSummaries, msgSummary{types.MsgSharesUsed(int(pTx.msg.MessageSize)), pTx.msg.MessageNameSpaceId})
	}

	txShares = txBytes / consts.TxShareSize
	if txBytes > 0 {
		txShares++ // add one to round up
	}

	for _, e := range evd.Evidence {
		evdBytes += e.Size() + types.DelimLen(uint64(e.Size()))
	}

	evdShares = evdBytes / consts.TxShareSize
	if evdBytes > 0 {
		evdShares++ // add one to round up
	}

	// sort the msgSummaries in order to order properly. This is okay to do here
	// as we aren't sorting the actual txs, just their summaries for more
	// accurate estimations
	sort.Slice(msgSummaries, func(i, j int) bool {
		return bytes.Compare(msgSummaries[i].namespace, msgSummaries[j].namespace) < 0
	})

	// isolate the sizes as we no longer need the namespaces
	msgShares := make([]int, len(msgSummaries))
	for i, summary := range msgSummaries {
		msgShares[i] = summary.size
	}

	return txShares, evdShares, msgShares
}

func nextPowerOfTwo(v int) int {
	k := 1
	for k < v {
		k = k << 1
	}
	return k
}
