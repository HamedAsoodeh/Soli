package shares

import coretypes "github.com/tendermint/tendermint/types"

// SplitMessageUsingNIDefaultsUnbounded needs a better name lol. splits the
// provided messages into shares following the non-interactive defaults. This is
// unbounded in that it doesn't actually stop splitting shares when the square
// is filled. This is useful for creating prioritized blocks. We can generate
// the a large number of shares, and then remove the least prioritized messages
// later until we have a set of shares that fit inside the block.
func SplitMessagesUsingNIDefaultsUnbounded(rowSize, start int, msgs []coretypes.Message) (*MessageShareSplitter, []uint32) {
	cursor := start
	indexes := []uint32{uint32(start)}
	splitter := NewMessageShareSplitter()
	for _, msg := range msgs {
		nextCursor, fits := NextAlignedPowerOfTwo(cursor, len(msg.Data), rowSize)
		// if the largest power of two portion of the message doesn't fit on
		// this row, then it must start on the next row.
		if !fits {
			nextCursor = ((cursor/rowSize)+1)*rowSize - 1
		}
		splitter.WriteNamespacedPaddedShares(nextCursor - cursor)
		splitter.Write(msg)
		indexes = append(indexes, uint32(nextCursor))
		cursor = nextCursor + len(msg.Data)
	}
	return splitter, indexes
}

// FitsInSquare uses the non interactive default rules to see if messages of
// some lengths will fit in a square of size origSquareSize starting at share
// index cursor. See non-interactive default rules
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#non-interactive-default-rules
func FitsInSquare(cursor, origSquareSize int, msgLens ...int) (fits bool) {
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

// NextAlignedPowerOfTwo calculates the next index in a row that is an aligned
// power of two or returns false is the msg cannot fit on the given row at the
// next aligned power of two. An aligned power of two means that the largest
// power of two that fits entirely in the msg or the square size. pls see specs
// for further details. Assumes that cursor < k, all args are non negative, and
// that k is a power of two.
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#non-interactive-default-rules
func NextAlignedPowerOfTwo(cursor, msgLen, k int) (int, bool) {
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
