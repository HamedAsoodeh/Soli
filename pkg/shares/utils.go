package shares

import (
	"math/bits"

	"github.com/tendermint/tendermint/pkg/consts"
)

// DelimLen calculates the length of the delimiter for a given message size
func DelimLen(x uint64) int {
	return 8 - bits.LeadingZeros64(x)%8
}

// MsgSharesUsed calculates the minimum number of shares a message will take up.
// It accounts for the necessary delimiter and potential padding.
func MsgSharesUsed(msgSize int) int {
	// add the delimiter to the message size
	msgSize = DelimLen(uint64(msgSize)) + msgSize
	shareCount := msgSize / consts.MsgShareSize
	// increment the share count if the message overflows the last counted share
	if msgSize%consts.MsgShareSize != 0 {
		shareCount++
	}
	return shareCount
}
