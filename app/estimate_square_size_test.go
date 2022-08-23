package app

import (
	"testing"

	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/pkg/consts"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
)

func Test_estimateSquareSize(t *testing.T) {
	type test struct {
		name                  string
		wPFDCount, messgeSize int
		expectedSize          uint64
	}
	tests := []test{
		{"empty block minimum square size", 0, 0, consts.MinSquareSize},
		{"random small block square size 2", 1, 400, 2},
		{"random small block square size 4", 1, 2000, 4},
		{"random small block square size 16", 4, 2000, 16},
		{"random medium block square size 32", 50, 2000, 32},
		{"full block max square size", 8000, 100, consts.MaxSquareSize},
		{"overly full block", 8000, 1000, consts.MaxSquareSize},
	}
	encConf := encoding.MakeConfig(ModuleEncodingRegisters...)
	signer := generateKeyringSigner(t, "estimate-key")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs := generateManyRawWirePFD(t, encConf.TxConfig, signer, tt.wPFDCount, tt.messgeSize)
			parsedTxs := parseTxs(encConf.TxConfig, txs)
			res, _ := estimateSquareSize(parsedTxs, core.EvidenceList{})
			assert.Equal(t, tt.expectedSize, res)
		})
	}
}

func Test_contiguousShareCount(t *testing.T) {
	type test struct {
		name                  string
		wPFDCount, messgeSize int
		expected              int
	}
	// todo: add other tx types
	tests := []test{
		{"empty block minimum square size", 0, 0, consts.MinSquareSize},
		{"random small block square size 2", 1, 400, 2},
		{"random small block square size 4", 1, 2000, 4},
		{"random small block square size 4", 4, 2000, 8},
		{"random medium block square size 32", 50, 2000, 32},
	}
	encConf := encoding.MakeConfig(ModuleEncodingRegisters...)
	signer := generateKeyringSigner(t, "estimate-key")
	for _, tt := range tests {
		txs := generateManyRawWirePFD(t, encConf.TxConfig, signer, tt.wPFDCount, tt.messgeSize)
		parsedTxs := parseTxs(encConf.TxConfig, txs)
		// malleate any malleable txs while also keeping track of the original order
		// and tagging the resulting messages with a reverse index.
		var err error
		rawTxs := make([][]byte, len(parsedTxs))
		for i, pTx := range parsedTxs {
			if pTx.msg != nil {
				err = pTx.malleate(encConf.TxConfig, 32)
				require.NoError(t, err)
				wt, err := pTx.wrap(1)
				require.NoError(t, err)
				rawTxs[i] = wt
			} else {
				rawTxs[i] = pTx.rawTx
			}
		}
		calculatedTxShareCount := calculateContigShareCount(parsedTxs, core.EvidenceList{})
		// wrapped, err := parsedTxs.wrap([]uint32{1, 2, 3, 4, 5})
		// require.NoError(t, err)
		txShares := shares.SplitTxs(shares.TxsFromBytes(rawTxs))
		assert.Equal(t, len(txShares), calculatedTxShareCount, tt.name)
	}

}
