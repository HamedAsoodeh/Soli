package app

import (
	"bytes"
	"sort"

	"github.com/celestiaorg/celestia-app/pkg/inclusion/appconsts"
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
		fits := checkFitInSquare(txShares+evdShares, squareSize, msgLens...)
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

// checkFitInSquare uses the non interactive default rules to see if messages of
// some lengths will fit in a square of size origSquareSize starting at share
// index cursor. See non-interactive default rules
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#non-interactive-default-rules
func checkFitInSquare(cursor, origSquareSize int, msgLens ...int) (fits bool) {
	// if there are 0 messages and the cursor already fits inside the square,
	// then we already know that everything fits in the square.
	if len(msgLens) == 0 && cursor/origSquareSize <= origSquareSize {
		return true
	}
	// iterate through all of the messages and apply the non-interactive default
	// rules to check if they will fit
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
	// perform one last check that catches the edge case where the last message
	// overflows rows
	return cursor/origSquareSize <= origSquareSize
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
