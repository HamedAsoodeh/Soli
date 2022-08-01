package app

import "testing"

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
