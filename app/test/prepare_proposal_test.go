package app_test

import (
	"bytes"
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/testutil"
	"github.com/celestiaorg/celestia-app/testutil/testtxs"
	"github.com/celestiaorg/nmt/namespace"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/pkg/consts"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPrepareProposal(t *testing.T) {
	signer := testutil.GenerateKeyringSigner(t, testtxs.TestAccountName)

	encCfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	testApp := testutil.SetupTestAppWithGenesisValSet(t)

	type test struct {
		input            abci.RequestPrepareProposal
		expectedMessages []*core.Message
		expectedTxs      int
	}

	firstNS := []byte{2, 2, 2, 2, 2, 2, 2, 2}
	firstMessage := bytes.Repeat([]byte{4}, 512)
	firstRawTx := testtxs.GenerateRawWirePFDTx(t, encCfg.TxConfig, firstNS, firstMessage, signer, 2, 4, 8, 16)

	secondNS := []byte{1, 1, 1, 1, 1, 1, 1, 1}
	secondMessage := []byte{2}
	secondRawTx := testtxs.GenerateRawWirePFDTx(t, encCfg.TxConfig, secondNS, secondMessage, signer, 2, 4, 8, 16)

	thirdNS := []byte{3, 3, 3, 3, 3, 3, 3, 3}
	thirdMessage := []byte{1}
	thirdRawTx := testtxs.GenerateRawWirePFDTx(t, encCfg.TxConfig, thirdNS, thirdMessage, signer, 2, 4, 8, 16)

	tests := []test{
		{
			input: abci.RequestPrepareProposal{
				BlockData: &core.Data{
					Txs: [][]byte{firstRawTx, secondRawTx, thirdRawTx},
				},
			},
			expectedMessages: []*core.Message{
				{
					NamespaceId: secondNS, // the second message should be first
					Data:        []byte{2},
				},
				{
					NamespaceId: firstNS,
					Data:        firstMessage,
				},
				{
					NamespaceId: thirdNS,
					Data:        []byte{1},
				},
			},
			expectedTxs: 3,
		},
	}

	for _, tt := range tests {
		res := testApp.PrepareProposal(tt.input)
		assert.Equal(t, tt.expectedMessages, res.BlockData.Messages.MessagesList)
		assert.Equal(t, tt.expectedTxs, len(res.BlockData.Txs))

		// verify the signatures of the prepared txs
		sdata, err := signer.GetSignerData()
		if err != nil {
			require.NoError(t, err)
		}
		dec := encoding.MalleatedTxDecoder(encCfg.TxConfig.TxDecoder())
		for _, tx := range res.BlockData.Txs {
			sTx, err := dec(tx)
			require.NoError(t, err)

			sigTx, ok := sTx.(authsigning.SigVerifiableTx)
			require.True(t, ok)

			sigs, err := sigTx.GetSignaturesV2()
			require.NoError(t, err)
			require.Equal(t, 1, len(sigs))
			sig := sigs[0]

			err = authsigning.VerifySignature(
				sdata.PubKey,
				sdata,
				sig.Data,
				encCfg.TxConfig.SignModeHandler(),
				sTx,
			)
			assert.NoError(t, err)
		}
	}
}

func TestPrepareMessagesWithReservedNamespaces(t *testing.T) {
	testApp := testutil.SetupTestAppWithGenesisValSet(t)
	encCfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	signer := testutil.GenerateKeyringSigner(t, testtxs.TestAccountName)

	type test struct {
		name             string
		namespace        namespace.ID
		expectedMessages int
	}

	tests := []test{
		{"transaction namespace id for message", consts.TxNamespaceID, 0},
		{"evidence namespace id for message", consts.EvidenceNamespaceID, 0},
		{"tail padding namespace id for message", consts.TailPaddingNamespaceID, 0},
		{"parity shares namespace id for message", consts.ParitySharesNamespaceID, 0},
		{"reserved namespace id for message", namespace.ID{0, 0, 0, 0, 0, 0, 0, 200}, 0},
		{"valid namespace id for message", namespace.ID{3, 3, 2, 2, 2, 1, 1, 1}, 1},
	}

	for _, tt := range tests {
		tx := testtxs.GenerateRawWirePFDTx(t, encCfg.TxConfig, tt.namespace, []byte{1}, signer, 2, 4, 8, 16)
		input := abci.RequestPrepareProposal{
			BlockData: &core.Data{
				Txs: [][]byte{tx},
			},
		}
		res := testApp.PrepareProposal(input)
		assert.Equal(t, tt.expectedMessages, len(res.BlockData.Messages.MessagesList))
	}
}
