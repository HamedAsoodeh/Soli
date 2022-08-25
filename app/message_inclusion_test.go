package app

import (
	"fmt"
	"math"
	"testing"

	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/inclusion"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/rsmt2d"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/pkg/consts"
	"github.com/tendermint/tendermint/pkg/da"
)

func TestMessageInclusion(t *testing.T) {
	type test struct {
		pfdCount, normalTxCount, size int
	}
	tests := []test{
		{1, 0, 100},
		{2, 0, 100},
		{1, 0, 300},
		{1, 0, 2000},
		{2, 0, 600},
		{10, 0, 1000},
		{20, 0, 2000},
		{10, 0, 16134},
		{1, 0, 900000},
		{1, 1, 100},
		{2, 10, 100},
		{1, 10, 300},
		{1, 10, 2000},
		{2, 20, 600},
		{10, 20, 1000},
		{20, 10, 2000},
		{10, 10, 16134},
		{1, 10, 900000},
		{1, 0, 62244},
	}
	encConf := encoding.MakeConfig(ModuleEncodingRegisters...)
	signer := generateKeyringSigner(t, "msg-inclusion-key")
	for tti, tt := range tests {
		data, err := GenerateValidBlockData(t, encConf.TxConfig, signer, tt.pfdCount, tt.normalTxCount, tt.size)
		require.NoError(t, err)
		dataSquare, err := shares.Split(data)
		require.NoError(t, err)

		squareSize := uint64(math.Sqrt(float64(len(dataSquare))))

		cacher := inclusion.NewSubtreeCacher(squareSize)
		eds, err := rsmt2d.ComputeExtendedDataSquare(dataSquare, consts.DefaultCodec(), cacher.Constructor)
		require.NoError(t, err)

		dah := da.NewDataAvailabilityHeader(eds)

		indexes := shares.ExtractShareIndexes(data.Txs)

		pfds := []*types.MsgPayForData{}
		for _, tx := range data.Txs {
			dec := encoding.MalleatedTxDecoder(encConf.TxConfig.TxDecoder())
			tx, err := dec(tx)
			require.NoError(t, err)
			for _, m := range tx.GetMsgs() {
				pfd, ok := m.(*types.MsgPayForData)
				if !ok {
					continue
				}
				pfds = append(pfds, pfd)
			}
		}

		t.Run(fmt.Sprintf("test %d: pfd count %d size %d", tti, tt.pfdCount, tt.size), func(t *testing.T) {
			for i, indx := range indexes {
				msgSharesUsed := shares.MsgSharesUsed(len(data.Messages.MessagesList[i].Data))
				commit, err := inclusion.GetCommit(cacher, dah, int(indx), msgSharesUsed)
				require.NoError(t, err)
				assert.Equal(t, pfds[i].MessageShareCommitment, commit)
			}
		})

	}
}
