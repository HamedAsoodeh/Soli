package app

import (
	"errors"

	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
)

func malleateTxs(txConf client.TxConfig, squareSize uint64, txs parsedTxs) (parsedTxs, []core.Message, error) {
	var err error
	var msgs []core.Message
	for _, pTx := range txs {
		if pTx.malleatedTx != nil {
			err = pTx.malleate(txConf, squareSize)
			msgs = append(msgs, pTx.message())
		}
	}
	return txs, msgs, err
}

func (p *parsedTx) malleate(txConf client.TxConfig, squareSize uint64) error {
	if p.msg == nil || p.tx == nil {
		return errors.New("can only malleate a tx with a MsgWirePayForData")
	}

	// parse wire message and create a single message
	_, unsignedPFD, sig, err := types.ProcessWirePayForData(p.msg, squareSize)
	if err != nil {
		return err
	}

	// create the signed PayForData using the fees, gas limit, and sequence from
	// the original transaction, along with the appropriate signature.
	signedTx, err := types.BuildPayForDataTxFromWireTx(p.tx, txConf.NewTxBuilder(), sig, unsignedPFD)
	if err != nil {
		return err
	}

	rawProcessedTx, err := txConf.TxEncoder()(signedTx)
	if err != nil {
		return err
	}

	p.malleatedTx = rawProcessedTx
	return nil
}
