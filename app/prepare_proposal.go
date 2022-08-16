package app

import (
	"github.com/celestiaorg/celestia-app/pkg/shares"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/pkg/da"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"
)

// PrepareProposal fullfills the celestia-core version of the ABCI interface by
// preparing the proposal block data. The square size is determined by first
// estimating it via the size of the passed block data. Then the included
// MsgWirePayForData messages are malleated into MsgPayForData messages by
// separating the message and transaction that pays for that message. Lastly,
// this method generates the data root for the proposal block and passes it back
// to tendermint via the blockdata.
func (app *App) PrepareProposal(req abci.RequestPrepareProposal) abci.ResponsePrepareProposal {
	// parse the txs, extracting any MsgWirePayForData and performing basic
	// validation for each transaction. Invalid txs are ignored.
	parsedTxs := parseTxs(app.txConfig, req.BlockData.Txs)

	// estimate the square size. This estimation errors on the side of larger
	// squares but can only return values within the min and max square size.
	squareSize, totalSharesUsed := estimateSquareSize(parsedTxs, req.BlockData.Evidence)

	// the totalSharesUsed can be larger that the max number of shares if we
	// reach the max square size. In this case, we must prune the deprioritized
	// txs (and their messages if they're pfd txs).
	if totalSharesUsed > int(squareSize*squareSize) {
		parsedTxs = prune(app.txConfig, parsedTxs, totalSharesUsed, int(squareSize))
	}

	malleatedTxs, messages := malleateTxs(app.txConfig, squareSize, parsedTxs)

	contigousShareCount := calculateContigShareCount(malleatedTxs, req.BlockData.Evidence)
	msgShareCounts := shares.MessageShareCountsFromMessages(messages)

	// calculate the indexes that will be used for each message
	_, indexes := shares.MsgSharesUsedNIDefaults(contigousShareCount, int(squareSize), msgShareCounts...)

	// wrap the malleated txs with their message's starting position
	wrappedMalleatedTxs, err := malleatedTxs.export(indexes)
	if err != nil {
		// todo handle
	}

	blockData := core.Data{
		Txs:                wrappedMalleatedTxs,
		Evidence:           req.BlockData.Evidence,
		Messages:           core.Messages{MessagesList: messages},
		OriginalSquareSize: squareSize,
	}

	coreData, err := coretypes.DataFromProto(&blockData)
	if err != nil {
		// todo handle
		panic(err)
	}

	var evd coretypes.EvidenceData
	err = evd.FromProto(&req.BlockData.Evidence)
	if err != nil {
		panic(err)
	}

	dataSquare, err := shares.Split(coreData)
	if err != nil {
		// todo: handle this panic even tho it should never get hit.
		panic(err)
	}

	// encode the parsed transactions to the share format and create the
	// protobuf encoded equivalent version of this block data. MsgWirePayForData
	// txs are malleated into MsgPayForData txs during this process. The
	// malleated txs are wrapped with meta data to indicate their position in
	// the square and the hash of the original wire tx. When writing messages to
	// the data square, we follow the non-interactive default rules to ensure
	// that the malleated payment txs (MsgPayForData) have a corresponding data
	// blob (message) in the same square.

	// erasure the data square which we use to create the data root.
	eds, err := da.ExtendShares(squareSize, dataSquare)
	if err != nil {
		app.Logger().Error(
			"failure to erasure the data square while creating a proposal block",
			"error",
			err.Error(),
		)
		panic(err)
	}

	// create the new data root by creating merkle roots of each row and col of
	// the erasure data.
	dah := da.NewDataAvailabilityHeader(eds)
	// We use the block data struct to pass the square size and calculated data
	// root to tendermint.
	blockData.Hash = dah.Hash()
	blockData.OriginalSquareSize = squareSize

	// tendermint doesn't need to use any of the erasure data, as only the
	// protobuf encoded version of the block data is gossiped.
	return abci.ResponsePrepareProposal{
		BlockData: &blockData,
	}
}
