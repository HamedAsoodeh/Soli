package inclusion

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/pkg/da"
)

func GetCommit(cacher *EDSSubTreeRootCacher, dah da.DataAvailabilityHeader, start, msgShareLen int) ([]byte, error) {
	originalSquareSize := len(dah.RowsRoots) / 2
	if start+msgShareLen > originalSquareSize*originalSquareSize {
		return nil, errors.New("cannot get commit for message that doesn't fit in square")
	}
	fmt.Println("PATH INPUT", originalSquareSize, start, msgShareLen)
	paths := calculateCommitPaths(originalSquareSize, start, msgShareLen)
	fmt.Println("pathsssssssssssssssss", paths)
	commits := make([][]byte, len(paths))
	for i, path := range paths {
		// here we prepend false (walk left down the tree) because we only need
		// the commits to the original square
		orignalSquarePath := append(append(make([]bool, 0, len(path.instructions)+1), false), path.instructions...)
		fmt.Println("original square paths", orignalSquarePath)
		commit, err := cacher.GetSubTreeRoot(dah, path.row, orignalSquarePath)
		if err != nil {
			return nil, err
		}
		commits[i] = commit

	}
	fmt.Println("when getting commit ---------------------***")
	fmt.Println("&^%", originalSquareSize, paths, commits)
	// todo fix
	commitOUt := merkle.HashFromByteSlices(commits)
	fmt.Println("commit out", commitOUt)
	return commitOUt, nil
}
