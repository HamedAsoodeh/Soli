package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_estimateSquareSize(t *testing.T) {
	type test struct {
		name         string
		Txs          [][]byte
		expectedSize int
	}
	tests := []test{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}

}

func Test_estimateNonInteractiveDefaultPadding(t *testing.T) {
	type test struct {
		name  string
		msgs  []int
		start int
		size  int
		fits  bool
	}
	tests := []test{
		{
			name:  "10 msgs size 10 shares (100 msg shares, 0 contiguous, size 4)",
			msgs:  []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
			start: 0,
			size:  4,
			fits:  false,
		},
		{
			name:  "15 msgs size 1 share (15 msg shares, 0 contiguous, size 4)",
			msgs:  []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			start: 0,
			size:  4,
			fits:  true,
		},
		{
			name:  "15 msgs size 1 share starting at share 2 (15 msg shares, 2 contiguous, size 4)",
			msgs:  []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			start: 2,
			size:  4,
			fits:  false,
		},
		{
			name:  "8 msgs of various sizes, 7 starting shares (63 msg shares, 1 contigous, size 8)",
			msgs:  []int{3, 9, 3, 7, 8, 3, 7, 8},
			start: 1,
			size:  8,
			fits:  true,
		},
		{
			name:  "8 msgs of various sizes, 7 starting shares (63 msg shares, 6 contigous, size 8)",
			msgs:  []int{3, 9, 3, 7, 8, 3, 7, 8},
			start: 6,
			size:  8,
			fits:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := estimateNonInteractiveDefaultPadding(tt.start, tt.size, tt.msgs)
			assert.Equal(t, tt.fits, res)
		})
	}
}
