package types

import (
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/tendermint/tendermint/pkg/consts"
)

// https://github.com/celestiaorg/celestia-app/issues/236
// https://github.com/celestiaorg/celestia-app/issues/239

var allSquareSizes = generateAllSquareSizes()

// generateAllSquareSizes generates and returns all of the possible square sizes
// using the maximum and minimum square sizes
func generateAllSquareSizes() []int {
	sizes := []int{}
	cursor := int(consts.MinSquareSize)
	for cursor <= consts.MaxSquareSize {
		sizes = append(sizes, cursor)
		cursor *= 2
	}
	return sizes
}

// AllSquareSizes calculates all of the square sizes that message could possibly
// fit in
func AllSquareSizes(msgSize int) []uint64 {
	allSizes := allSquareSizes
	fitSizes := []uint64{}
	shareCount := shares.MsgSharesUsed(msgSize)
	for _, size := range allSizes {
		// if the number of shares is larger than that in the square, throw an error
		// note, we use k*k-1 here because at least a single share will be reserved
		// for the transaction paying for the message, therefore the max number of
		// shares a message can be is number of shares in square -1.
		if shareCount > (size*size)-1 {
			continue
		}
		fitSizes = append(fitSizes, uint64(size))
	}
	return fitSizes
}
