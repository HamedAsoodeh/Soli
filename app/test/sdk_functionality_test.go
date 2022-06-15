package app_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/testutil/network"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmosnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	legacygovtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
)

// BasicFuncTestSuite is used to check for basic functionality from the cosmos-sdk
type BasicFuncTestSuite struct {
	suite.Suite

	cfg      cosmosnet.Config
	encCfg   encoding.EncodingConfig
	network  *cosmosnet.Network
	kr       keyring.Keyring
	accounts []string
}

func NewBasicFuncTestSuite(cfg cosmosnet.Config) *BasicFuncTestSuite {
	return &BasicFuncTestSuite{cfg: cfg}
}

func (s *BasicFuncTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	numAccounts := 100
	s.accounts = make([]string, numAccounts)
	for i := 0; i < numAccounts; i++ {
		s.accounts[i] = tmrand.Str(20)
	}

	net := network.New(s.T(), s.cfg, s.accounts...)

	err := network.GRPCConn(net)
	s.Require().NoError(err)
	s.network = net
	s.kr = net.Validators[0].ClientCtx.Keyring
	s.encCfg = encoding.MakeEncodingConfig(app.ModuleEncodingRegisters...)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *BasicFuncTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestBasicFuncTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.EnableTMLogging = false
	cfg.MinGasPrices = "0utia"
	cfg.NumValidators = 2
	suite.Run(t, NewBasicFuncTestSuite(cfg))
}

func (s *BasicFuncTestSuite) TestGovModule() {
	require := s.Require()
	assert := s.Assert()
	val := s.network.Validators[0]
	govModuleAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	type test struct {
		name   string
		msgGen func(client.Context) sdk.Msg
	}

	tests := []test{
		{
			"submit a legacy text governace proposal",
			func(c client.Context) sdk.Msg {
				kr := c.Keyring

				rec, err := kr.Key(s.accounts[0])
				require.NoError(err)

				addr, err := rec.GetAddress()
				require.NoError(err)

				coins := sdk.NewCoins(
					sdk.NewCoin(app.BondDenom, sdk.NewInt(1000000000)),
				)
				propContent := legacygovtypes.NewTextProposal("test", "anarchy")
				msgContent, err := v1.NewLegacyContent(propContent, govModuleAddress)
				require.NoError(err)

				msg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgContent}, coins, addr.String(), "none")
				require.NoError(err)

				return msg
			},
		},
		{
			"submit a legacy params change",
			func(c client.Context) sdk.Msg {
				jsonProposal := `{
					"title": "Increase Signed Blocks Window Parameter to 2880",
					"description": "Mamaki Testnet initially started with very strict slashing conditions. This proposal changes the signed_blocks_window to about 24 hours.",
					"changes": [
					  {
						"subspace": "slashing",
						"key": "SignedBlocksWindow",
						"value": "2880"
					  }
					],
					"deposit": "100000utia"
				  }`

				kr := c.Keyring

				rec, err := kr.Key(s.accounts[0])
				require.NoError(err)

				addr, err := rec.GetAddress()
				require.NoError(err)
				var proposal paramscutils.ParamChangeProposalJSON
				err = json.Unmarshal([]byte(jsonProposal), &proposal)
				require.NoError(err)

				content := paramproposal.NewParameterChangeProposal(
					proposal.Title, proposal.Description, proposal.Changes.ToParamChanges(),
				)

				deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
				require.NoError(err)

				msg, err := legacygovtypes.NewMsgSubmitProposal(content, deposit, addr)
				require.NoError(err)

				return msg
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx
			kr := clientCtx.Keyring
			node, err := clientCtx.GetNode()
			require.NoError(err)

			// quick check of balances
			bals, err := queryForBalance(clientCtx, s.accounts[0])
			require.NoError(err)
			fmt.Println(bals)

			signer := types.NewKeyringSigner(kr, s.accounts[0], clientCtx.ChainID)

			err = signer.UpdateAccountFromClient(clientCtx)
			require.NoError(err)

			coin := sdk.Coin{
				Denom:  app.BondDenom,
				Amount: sdk.NewInt(100000000),
			}

			opts := []types.TxBuilderOption{
				types.SetFeeAmount(sdk.NewCoins(coin)),
				types.SetGasLimit(1000000000),
			}

			builder := signer.NewTxBuilder(opts...)

			msg := tc.msgGen(clientCtx)

			tx, err := signer.BuildSignedTx(builder, msg)
			require.NoError(err)

			rawTx, err := s.cfg.TxConfig.TxEncoder()(tx)
			require.NoError(err)

			rec := signer.GetSignerInfo()
			addr, err := rec.GetAddress()
			require.NoError(err)

			res, err := val.ClientCtx.BroadcastTxSync(rawTx)
			require.NoError(err)
			fmt.Println("sync resp", res.Logs, res.RawLog, res.Info, "signer", addr.String())
			fmt.Println("granter", tx.FeeGranter().String(), "payer", tx.FeePayer().String())
			assert.Equal(abci.CodeTypeOK, res.Code)
			hexHash := res.TxHash

			// wait a block to clear the txs
			require.NoError(s.network.WaitForNextBlock())
			require.NoError(s.network.WaitForNextBlock())

			hash, err := hex.DecodeString(hexHash)
			require.NoError(err)

			qres, err := node.Tx(context.Background(), hash, false)
			require.NoError(err)

			fmt.Println("query", qres.TxResult.Code, "LOGG", qres.TxResult.Log, "EVENTS", qres.TxResult.Events, "INFO", qres.TxResult.Info)
			assert.Greater(qres.TxResult.GasUsed, int64(0))
		})
	}
}

func (s *BasicFuncTestSuite) TestBankModule() {
	require := s.Require()
	assert := s.Assert()
	val := s.network.Validators[0]

	type test struct {
		name   string
		msgGen func(client.Context) sdk.Msg
	}

	tests := []test{
		{
			"submit a text governace proposal",
			func(c client.Context) sdk.Msg {
				msg, err := createSendMsg(c, s.accounts[0], s.accounts[1], 1000000)
				require.NoError(err)
				return msg
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx
			kr := clientCtx.Keyring
			node, err := clientCtx.GetNode()
			require.NoError(err)

			// quick check of balances
			bals, err := queryForBalance(clientCtx, s.accounts[0])
			require.NoError(err)
			fmt.Println(bals)

			signer := types.NewKeyringSigner(kr, s.accounts[0], clientCtx.ChainID)

			err = signer.UpdateAccountFromClient(clientCtx)
			require.NoError(err)

			coin := sdk.Coin{
				Denom:  app.BondDenom,
				Amount: sdk.NewInt(1000000),
			}

			opts := []types.TxBuilderOption{
				types.SetFeeAmount(sdk.NewCoins(coin)),
				types.SetGasLimit(1000000000),
			}

			builder := signer.NewTxBuilder(opts...)

			msg := tc.msgGen(clientCtx)

			tx, err := signer.BuildSignedTx(builder, msg)
			require.NoError(err)

			rawTx, err := s.cfg.TxConfig.TxEncoder()(tx)
			require.NoError(err)

			res, err := val.ClientCtx.BroadcastTxSync(rawTx)
			require.NoError(err)
			assert.Equal(abci.CodeTypeOK, res.Code)
			hexHash := res.TxHash

			// wait a block to clear the txs
			require.NoError(s.network.WaitForNextBlock())
			require.NoError(s.network.WaitForNextBlock())

			hash, err := hex.DecodeString(hexHash)
			require.NoError(err)

			qres, err := node.Tx(context.Background(), hash, false)
			require.NoError(err)

			fmt.Println(qres.TxResult.Code, qres.TxResult.Log, qres.TxResult.GasUsed, qres.TxResult.GasWanted)
		})
	}
}

func queryForBalance(c client.Context, acc string) (string, error) {
	kr := c.Keyring
	rec, err := kr.Key(acc)
	if err != nil {
		return "", err
	}

	addr, err := rec.GetAddress()
	if err != nil {
		return "", err
	}

	qc := banktypes.NewQueryClient(c.GRPCClient)
	res, err := qc.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: addr.String(),
	})
	if err != nil {
		return "", err
	}

	return res.Balances.String(), nil
}

func createSendMsg(c client.Context, acc1, acc2 string, amount int64) (sdk.Msg, error) {
	kr := c.Keyring
	rec1, err := kr.Key(acc1)
	if err != nil {
		return nil, err
	}
	addr1, err := rec1.GetAddress()
	if err != nil {
		return nil, err
	}
	rec2, err := kr.Key(acc2)
	if err != nil {
		return nil, err
	}
	addr2, err := rec2.GetAddress()
	if err != nil {
		return nil, err
	}
	coins := sdk.NewCoins(
		sdk.NewCoin(app.BondDenom, sdk.NewInt(amount)),
	)
	return banktypes.NewMsgSend(addr1, addr2, coins), nil
}
