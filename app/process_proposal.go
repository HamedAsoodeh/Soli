package app

import (
	"bytes"
	"fmt"

	"github.com/celestiaorg/celestia-app/pkg/inclusion"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/rsmt2d"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/pkg/consts"
	"github.com/tendermint/tendermint/pkg/da"
	coretypes "github.com/tendermint/tendermint/types"
)

const (
	rejectedPropBlockLog = "Rejected proposal block:"
)

func (app *App) ProcessProposal(req abci.RequestProcessProposal) abci.ResponseProcessProposal {
	// Check for message inclusion:
	//  - each MsgPayForData included in a block should have a corresponding data also in the block body
	//  - the commitment in each PFD should match that of its corresponding data
	//  - there should be no unpaid-for data

	data, err := coretypes.DataFromProto(req.BlockData)
	if err != nil {
		app.Logger().Error(rejectedPropBlockLog, "reason", "failure to unmarshal block data:", "error", err)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	fmt.Println("messages compared to txs", len(data.Messages.MessagesList), len(data.Txs))

	dataSquare, err := shares.Split(data)
	if err != nil {
		app.Logger().Error(rejectedPropBlockLog, "reason", "failure to compute shares from block data:", "error", err)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	cacher := inclusion.NewSubtreeCacher(data.OriginalSquareSize)
	eds, err := rsmt2d.ComputeExtendedDataSquare(dataSquare, consts.DefaultCodec(), cacher.Constructor)
	if err != nil {
		app.Logger().Error(
			rejectedPropBlockLog,
			"reason",
			"failure to erasure the data square",
			"error",
			err,
		)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	dah := da.NewDataAvailabilityHeader(eds)

	if !bytes.Equal(dah.Hash(), req.Header.DataHash) {
		app.Logger().Error(
			rejectedPropBlockLog,
			"reason",
			"proposed data root differs from calculated data root",
		)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	// iterate over all of the MsgPayForData transactions and ensure that they
	// commitments are subtree roots of the data root.
	commitmentCounter := 0
	for _, rawTx := range req.BlockData.Txs {
		malleatedTx, isMalleated := coretypes.UnwrapMalleatedTx(rawTx)
		if !isMalleated {
			fmt.Println("tx was not malleated")
			continue
		}

		tx, err := app.txConfig.TxDecoder()(malleatedTx.Tx)
		if err != nil {
			// we don't reject the block here because it is not a block validity
			// rule that all transactions included in the block data are
			// decodable
			fmt.Println("could not decode", err)
			continue
		}

		for _, msg := range tx.GetMsgs() {
			if sdk.MsgTypeURL(msg) != types.URLMsgPayForData {
				fmt.Println("msg was not a pay for data")
				continue
			}

			pfd, ok := msg.(*types.MsgPayForData)
			if !ok {
				app.Logger().Error("Msg type does not match MsgPayForData URL")
				continue
			}

			if err = pfd.ValidateBasic(); err != nil {
				app.Logger().Error(
					rejectedPropBlockLog,
					"reason",
					"invalid MsgPayForData",
					"error",
					err.Error(),
				)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			commitment, err := inclusion.GetCommit(cacher, dah, int(malleatedTx.ShareIndex), shares.MsgSharesUsed(int(pfd.MessageSize)))
			if err != nil {
				fmt.Println("commitment not found", err, "shares used", shares.MsgSharesUsed(int(pfd.MessageSize)))
				app.Logger().Error(
					rejectedPropBlockLog,
					"reason",
					"commitment not found",
					"error",
					err.Error(),
				)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			if !bytes.Equal(pfd.MessageShareCommitment, commitment) {
				fmt.Println("pfd and commitment do not match", "shares used", shares.MsgSharesUsed(int(pfd.MessageSize)), pfd.MessageSize)
				// todo: create a message inclusion proof
				app.Logger().Error(
					rejectedPropBlockLog,
					"reason",
					"found commitment does not match user's",
				)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			commitmentCounter++
		}
	}

	// compare the number of PFDs and messages, if they aren't
	// identical, then  we already know this block is invalid
	if commitmentCounter != len(req.BlockData.Messages.MessagesList) {
		app.Logger().Error(
			rejectedPropBlockLog,
			"reason",
			"varying number of messages and payForData txs in the same block",
		)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	fmt.Println("ACCEPTING PROPOSAL")
	return abci.ResponseProcessProposal{
		Result: abci.ResponseProcessProposal_ACCEPT,
	}
}
