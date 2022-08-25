package inclusion

import (
	"errors"

	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/pkg/da"
)

func GetCommit(cacher *EDSSubTreeRootCacher, dah da.DataAvailabilityHeader, start, msgShareLen int) ([]byte, error) {
	originalSquareSize := len(dah.RowsRoots) / 2
	if start+msgShareLen > originalSquareSize*originalSquareSize {
		return nil, errors.New("cannot get commit for message that doesn't fit in square")
	}
	paths := calculateCommitPaths(originalSquareSize, start, msgShareLen)
	commits := make([][]byte, len(paths))
	for i, path := range paths {
		// here we prepend false (walk left down the tree) because we only need
		// the commits to the original square
		orignalSquarePath := append(append(make([]bool, 0, len(path.instructions)+1), false), path.instructions...)
		commit, err := cacher.GetSubTreeRoot(dah, path.row, orignalSquarePath)
		if err != nil {
			return nil, err
		}
		commits[i] = commit

	}
	return merkle.HashFromByteSlices(commits), nil
}
