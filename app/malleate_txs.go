package app

import (
	"crypto/sha256"
	"errors"

	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"
)

// malleateTx performs the malleation process which extracts the message and
// metadata contained in the MsgWirePayForData to create a new transaction that
// contains a MsgPayForData.
func malleateTx(
	txConf client.TxConfig,
	tx parsedTx,
	squareSize uint64,
	shareIndex uint32,
) (malleatedTx coretypes.Tx, msg *core.Message, err error) {
	if tx.msg == nil || tx.tx == nil {
		return nil, nil, errors.New("can only malleate a tx with a MsgWirePayForData")
	}

	// parse wire message and create a single message
	coreMsg, unsignedPFD, sig, err := types.ProcessWirePayForData(tx.msg, squareSize)
	if err != nil {
		return nil, nil, err
	}

	// create the signed PayForData using the fees, gas limit, and sequence from
	// the original transaction, along with the appropriate signature.
	signedTx, err := types.BuildPayForDataTxFromWireTx(tx.tx, txConf.NewTxBuilder(), sig, unsignedPFD)
	if err != nil {
		return nil, nil, err
	}

	rawProcessedTx, err := txConf.TxEncoder()(signedTx)
	if err != nil {
		return nil, nil, err
	}

	originalHash := sha256.Sum256(tx.rawTx)

	// TODO: pass the share index when we start using a branch of tendermint that supports wrapped txs
	wrappedTx, err := coretypes.WrapMalleatedTx(originalHash[:], rawProcessedTx)
	if err != nil {
		return nil, nil, err
	}

	return wrappedTx, coreMsg, nil
}
