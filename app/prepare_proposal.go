package app

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/pkg/da"
)

// PrepareProposal fullfills the celestia-core version of the ABCI interface by
// preparing the proposal block data. The square size is determined by first
// estimating it via the size of the passed block data. Then the included
// MsgWirePayForData messages are malleated into MsgPayForData messages by
// separating the message and transaction that pays for that message. Lastly,
// this method generates the data root for the proposal block and passes it the
// blockdata.
func (app *App) PrepareProposal(req abci.RequestPrepareProposal) abci.ResponsePrepareProposal {
	squareSize := estimateSquareSize(req.BlockData, app.txConfig)

	dataSquare, data := SplitShares(app.txConfig, squareSize, req.BlockData)

	eds, err := da.ExtendShares(squareSize, dataSquare)
	if err != nil {
		app.Logger().Error(
			"failure to erasure the data square while creating a proposal block",
			"error",
			err.Error(),
		)
		panic(err)
	}

	dah := da.NewDataAvailabilityHeader(eds)
	data.Hash = dah.Hash()
	data.OriginalSquareSize = squareSize

	return abci.ResponsePrepareProposal{
		BlockData: data,
	}
}
