package app_test

import (
	"crypto/rand"
	"crypto/sha256"
	"math"
	"math/big"
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/inclusion"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/testutil"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/nmt/namespace"
	"github.com/celestiaorg/rsmt2d"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/pkg/consts"
	"github.com/tendermint/tendermint/pkg/da"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"
)

// todo: fix or delete as this test is kinda redundant
// func TestMessageInclusionCheck(t *testing.T) {
// 	signer := testutil.GenerateKeyringSigner(t, testtxs.TestAccountName)

// 	testApp := testutil.SetupTestAppWithGenesisValSet(t)

// 	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

// 	firstValidPFD, msg1 := genRandMsgPayForData(t, signer, 4)
// 	secondValidPFD, msg2 := genRandMsgPayForData(t, signer, 4)

// 	invalidCommitmentPFD, msg3 := genRandMsgPayForData(t, signer, 4)
// 	invalidCommitmentPFD.MessageShareCommitment = tmrand.Bytes(32)

// 	// block with all messages included
// 	validData := core.Data{
// 		Txs: [][]byte{
// 			buildTx(t, signer, encConf.TxConfig, firstValidPFD, 8),
// 			buildTx(t, signer, encConf.TxConfig, secondValidPFD, 8),
// 		},
// 		Messages: core.Messages{
// 			MessagesList: []*core.Message{
// 				{
// 					NamespaceId: firstValidPFD.MessageNamespaceId,
// 					Data:        msg1,
// 				},
// 				{
// 					NamespaceId: secondValidPFD.MessageNamespaceId,
// 					Data:        msg2,
// 				},
// 			},
// 		},
// 		OriginalSquareSize: 8,
// 	}

// 	// block with a missing message
// 	missingMessageData := core.Data{
// 		Txs: [][]byte{
// 			buildTx(t, signer, encConf.TxConfig, firstValidPFD, 4),
// 			buildTx(t, signer, encConf.TxConfig, secondValidPFD, 4),
// 		},
// 		Messages: core.Messages{
// 			MessagesList: []*core.Message{
// 				{
// 					NamespaceId: firstValidPFD.MessageNamespaceId,
// 					Data:        msg1,
// 				},
// 			},
// 		},
// 		OriginalSquareSize: 4,
// 	}

// 	// block with all messages included, but the commitment is changed
// 	invalidData := core.Data{
// 		Txs: [][]byte{
// 			buildTx(t, signer, encConf.TxConfig, firstValidPFD, 4),
// 			buildTx(t, signer, encConf.TxConfig, secondValidPFD, 4),
// 		},
// 		Messages: core.Messages{
// 			MessagesList: []*core.Message{
// 				{
// 					NamespaceId: firstValidPFD.MessageNamespaceId,
// 					Data:        msg1,
// 				},
// 				{
// 					NamespaceId: invalidCommitmentPFD.MessageNamespaceId,
// 					Data:        msg3,
// 				},
// 			},
// 		},
// 		OriginalSquareSize: 4,
// 	}

// 	// block with all messages included
// 	extraMessageData := core.Data{
// 		Txs: [][]byte{
// 			buildTx(t, signer, encConf.TxConfig, firstValidPFD, 4),
// 		},
// 		Messages: core.Messages{
// 			MessagesList: []*core.Message{
// 				{
// 					NamespaceId: firstValidPFD.MessageNamespaceId,
// 					Data:        msg1,
// 				},
// 				{
// 					NamespaceId: secondValidPFD.MessageNamespaceId,
// 					Data:        msg2,
// 				},
// 			},
// 		},
// 		OriginalSquareSize: 4,
// 	}

// 	type test struct {
// 		input          abci.RequestProcessProposal
// 		expectedResult abci.ResponseProcessProposal_Result
// 	}

// 	tests := []test{
// 		{
// 			input: abci.RequestProcessProposal{
// 				BlockData: &validData,
// 			},
// 			expectedResult: abci.ResponseProcessProposal_ACCEPT,
// 		},
// 		{
// 			input: abci.RequestProcessProposal{
// 				BlockData: &missingMessageData,
// 			},
// 			expectedResult: abci.ResponseProcessProposal_REJECT,
// 		},
// 		{
// 			input: abci.RequestProcessProposal{
// 				BlockData: &invalidData,
// 			},
// 			expectedResult: abci.ResponseProcessProposal_REJECT,
// 		},
// 		{
// 			input: abci.RequestProcessProposal{
// 				BlockData: &extraMessageData,
// 			},
// 			expectedResult: abci.ResponseProcessProposal_REJECT,
// 		},
// 	}

// 	for _, tt := range tests {
// 		data, err := coretypes.DataFromProto(tt.input.BlockData)
// 		require.NoError(t, err)

// 		shares, err := shares.Split(data)
// 		require.NoError(t, err)

// 		require.NoError(t, err)
// 		eds, err := da.ExtendShares(tt.input.BlockData.OriginalSquareSize, shares)
// 		require.NoError(t, err)
// 		dah := da.NewDataAvailabilityHeader(eds)
// 		tt.input.Header.DataHash = dah.Hash()
// 		res := testApp.ProcessProposal(tt.input)
// 		assert.Equal(t, tt.expectedResult, res.Result)
// 	}
// }

// func TestProcessMessagesWithReservedNamespaces(t *testing.T) {
// 	testApp := testutil.SetupTestAppWithGenesisValSet(t)
// 	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

// 	signer := testutil.GenerateKeyringSigner(t, testtxs.TestAccountName)

// 	type test struct {
// 		name           string
// 		namespace      namespace.ID
// 		expectedResult abci.ResponseProcessProposal_Result
// 	}

// 	tests := []test{
// 		{"transaction namespace id for message", consts.TxNamespaceID, abci.ResponseProcessProposal_REJECT},
// 		{"evidence namespace id for message", consts.EvidenceNamespaceID, abci.ResponseProcessProposal_REJECT},
// 		{"tail padding namespace id for message", consts.TailPaddingNamespaceID, abci.ResponseProcessProposal_REJECT},
// 		{"namespace id 200 for message", namespace.ID{0, 0, 0, 0, 0, 0, 0, 200}, abci.ResponseProcessProposal_REJECT},
// 		{"correct namespace id for message", namespace.ID{3, 3, 2, 2, 2, 1, 1, 1}, abci.ResponseProcessProposal_ACCEPT},
// 	}

// 	for _, tt := range tests {
// 		pfd, msg := genRandMsgPayForDataForNamespace(t, signer, 8, tt.namespace)
// 		input := abci.RequestProcessProposal{
// 			BlockData: &core.Data{
// 				Txs: [][]byte{
// 					buildTx(t, signer, encConf.TxConfig, pfd, 4),
// 				},
// 				Messages: core.Messages{
// 					MessagesList: []*core.Message{
// 						{
// 							NamespaceId: pfd.GetMessageNamespaceId(),
// 							Data:        msg,
// 						},
// 					},
// 				},
// 				OriginalSquareSize: 8,
// 			},
// 		}
// 		data, err := coretypes.DataFromProto(input.BlockData)
// 		require.NoError(t, err)

// 		shares, err := shares.Split(data)
// 		require.NoError(t, err)

// 		require.NoError(t, err)
// 		eds, err := da.ExtendShares(input.BlockData.OriginalSquareSize, shares)
// 		require.NoError(t, err)
// 		dah := da.NewDataAvailabilityHeader(eds)
// 		input.Header.DataHash = dah.Hash()
// 		res := testApp.ProcessProposal(input)
// 		assert.Equal(t, tt.expectedResult, res.Result)
// 	}
// }

// func TestProcessMessageWithParityShareNamespaces(t *testing.T) {
// 	testApp := testutil.SetupTestAppWithGenesisValSet(t)
// 	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

// 	signer := testutil.GenerateKeyringSigner(t, testtxs.TestAccountName)

// 	pfd, msg := genRandMsgPayForDataForNamespace(t, signer, 8, consts.ParitySharesNamespaceID)
// 	input := abci.RequestProcessProposal{
// 		BlockData: &core.Data{
// 			Txs: [][]byte{
// 				buildTx(t, signer, encConf.TxConfig, pfd, 4),
// 			},
// 			Messages: core.Messages{
// 				MessagesList: []*core.Message{
// 					{
// 						NamespaceId: pfd.GetMessageNamespaceId(),
// 						Data:        msg,
// 					},
// 				},
// 			},
// 			OriginalSquareSize: 8,
// 		},
// 	}
// 	res := testApp.ProcessProposal(input)
// 	assert.Equal(t, abci.ResponseProcessProposal_REJECT, res.Result)
// }

func TestProcessProposalMessageInclusion(t *testing.T) {
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
	}
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	signer := testutil.GenerateKeyringSigner(t, "msg-inclusion-key")
	testApp := testutil.SetupTestAppWithGenesisValSet(t)
	for _, tt := range tests {
		data, err := app.GenerateValidBlockData(t, encConf.TxConfig, signer, tt.pfdCount, tt.normalTxCount, tt.size)
		require.NoError(t, err)
		dataSquare, err := shares.Split(data)
		require.NoError(t, err)
		protoData := data.ToProto()

		squareSize := uint64(math.Sqrt(float64(len(dataSquare))))

		cacher := inclusion.NewSubtreeCacher(squareSize)
		eds, err := rsmt2d.ComputeExtendedDataSquare(dataSquare, consts.DefaultCodec(), cacher.Constructor)
		require.NoError(t, err)

		dah := da.NewDataAvailabilityHeader(eds)

		i := abci.RequestProcessProposal{
			BlockData: &protoData,
			Header: core.Header{
				DataHash: dah.Hash(),
			},
		}

		res := testApp.ProcessProposal(i)
		require.Equal(t, abci.ResponseProcessProposal_ACCEPT, res.Result)
	}
}

func genRandMsgPayForData(t *testing.T, signer *types.KeyringSigner, squareSize uint64) (*types.MsgPayForData, []byte) {
	ns := make([]byte, consts.NamespaceSize)
	_, err := rand.Read(ns)
	require.NoError(t, err)
	return genRandMsgPayForDataForNamespace(t, signer, squareSize, ns)
}

func genRandMsgPayForDataForNamespace(t *testing.T, signer *types.KeyringSigner, squareSize uint64, ns namespace.ID) (*types.MsgPayForData, []byte) {
	message := make([]byte, randomInt(20))
	_, err := rand.Read(message)
	require.NoError(t, err)

	commit, err := types.CreateCommitment(squareSize, ns, message)
	require.NoError(t, err)

	pfd := types.MsgPayForData{
		MessageShareCommitment: commit,
		MessageNamespaceId:     ns,
	}

	return &pfd, message
}

func buildTx(t *testing.T, signer *types.KeyringSigner, txCfg client.TxConfig, msg sdk.Msg, shareIndex uint32) []byte {
	tx, err := signer.BuildSignedTx(signer.NewTxBuilder(), msg)
	require.NoError(t, err)

	rawTx, err := txCfg.TxEncoder()(tx)
	require.NoError(t, err)

	h := sha256.Sum256(rawTx)

	coretypes.WrapMalleatedTx(h[:], shareIndex, rawTx)

	return rawTx
}

func randomInt(max int64) int64 {
	i, _ := rand.Int(rand.Reader, big.NewInt(max))
	return i.Int64()
}
