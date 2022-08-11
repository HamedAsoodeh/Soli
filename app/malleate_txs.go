package app

import (
	"crypto/sha256"
	"errors"

	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	coretypes "github.com/tendermint/tendermint/types"
)

type malleatedTx struct {
	malleatedTx  coretypes.Tx
	originalHash []byte
	namespace    []byte
	msg          coretypes.Message
}

// malleateTxs separates and malleates a set of parsed transactions.
func malleateTxs(txConf client.TxConfig, squareSize uint64, txs []parsedTx) ([]coretypes.Tx, []malleatedTx, error) {
	var mTxs []malleatedTx
	var normalTxs []coretypes.Tx
	for _, tx := range txs {
		if tx.msg == nil {
			normalTxs = append(normalTxs, tx.rawTx)
			continue
		}
		mTx, err := malleateTx(txConf, squareSize, tx)
		if err != nil {
			return nil, nil, err
		}
		mTxs = append(mTxs, mTx)
	}
	return normalTxs, mTxs, nil
}

// malleateTx performs the malleation process which extracts the message and
// metadata contained in the MsgWirePayForData to create a new transaction that
// contains a MsgPayForData.
func malleateTx(txConf client.TxConfig, squareSize uint64, tx parsedTx) (mTx malleatedTx, err error) {
	if tx.msg == nil || tx.tx == nil {
		return mTx, errors.New("can only malleate a tx with a MsgWirePayForData")
	}

	// parse wire message and create a single message
	coreMsg, unsignedPFD, sig, err := types.ProcessWirePayForData(tx.msg, squareSize)
	if err != nil {
		return mTx, err
	}

	// create the signed PayForData using the fees, gas limit, and sequence from
	// the original transaction, along with the appropriate signature.
	signedTx, err := types.BuildPayForDataTxFromWireTx(tx.tx, txConf.NewTxBuilder(), sig, unsignedPFD)
	if err != nil {
		return mTx, err
	}

	rawProcessedTx, err := txConf.TxEncoder()(signedTx)
	if err != nil {
		return mTx, err
	}

	originalHash := sha256.Sum256(tx.rawTx)

	return malleatedTx{
		malleatedTx:  rawProcessedTx,
		originalHash: originalHash[:],
		msg:          coreMsg,
	}, nil

}
