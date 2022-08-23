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
		pfdCount, size int
	}
	tests := []test{
		// {1, 100},
		{2, 100},
		// {1, 300},
		// {1, 2000},
		// {32, 10000},
	}
	encConf := encoding.MakeConfig(ModuleEncodingRegisters...)
	signer := generateKeyringSigner(t, "msg-inclusion-key")
	for tti, tt := range tests {
		data, err := generateValidBlockData(t, encConf.TxConfig, signer, tt.pfdCount, tt.size)
		require.NoError(t, err)
		dataSquare, err := shares.Split(data)
		require.NoError(t, err)

		squareSize := uint64(math.Sqrt(float64(len(dataSquare))))

		cacher := inclusion.NewSubtreeCacher(squareSize)
		eds, err := rsmt2d.ComputeExtendedDataSquare(dataSquare, consts.DefaultCodec(), cacher.Constructor)
		require.NoError(t, err)

		dah := da.NewDataAvailabilityHeader(eds)
		fmt.Println("message used according to shits", shares.MsgSharesUsed(len(data.Messages.MessagesList[0].Data)))
		for i, r := range eds.Row(0) {
			fmt.Println(i, r[:45])
		}

		indexes := shares.ExtractShareIndexes(data.Txs)
		fmt.Println("indexes ^^^^^^^^^^^^^^6", indexes)

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

		fmt.Println("eds eds eds eds eds eds")
		// fmt.Println(eds.Row(1)[0])
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

		t.Run(fmt.Sprintf("test %d: pfd count %d size %d", tti, tt.pfdCount, tt.size), func(t *testing.T) {
			for i, indx := range indexes {
				msgSharesUsed := shares.MsgSharesUsed(len(data.Messages.MessagesList[i].Data))
				fmt.Println("shares used during test", msgSharesUsed)
				commit, err := inclusion.GetCommit(cacher, dah, int(indx), msgSharesUsed)
				require.NoError(t, err)
				fmt.Println("comparing", pfds[i].MessageShareCommitment, commit)
				assert.Equal(t, pfds[i].MessageShareCommitment, commit)
			}
		})

	}
}
