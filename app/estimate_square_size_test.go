package app

import (
	"testing"

	"github.com/tendermint/tendermint/pkg/consts"
)

func TestEstimateSquareSize(t *testing.T) {
	type test struct {
		name                  string
		wPFDCount, messgeSize int
		expectedSize          uint64
	}
	tests := []test{
		{"empty block minimum square size", 0, 0, consts.MinSquareSize},
		{"random small block square size 2", 1, 400, 2},
		{"random small block square size 4", 1, 2000, 4},
		{"random small block square size 4", 4, 2000, 8},
		{"random medium block square size 32", 50, 2000, 32},
		{"full block max square size", 16000, 200, consts.MaxSquareSize},
		{"overly full block", 16000, 1000, consts.MaxSquareSize},
	}
	// encConf := encoding.MakeConfig(ModuleEncodingRegisters...)
	// signer := generateKeyringSigner(t, "estimate-key")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// txs := generateManyRawWirePFD(t, encConf.TxConfig, signer, tt.wPFDCount, tt.messgeSize)
			// res := estimateSquareSize(&core.Data{Txs: txs}, encConf.TxConfig)
			// assert.Equal(t, tt.expectedSize, res)
		})
	}

}
