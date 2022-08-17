package app

import (
	"errors"

	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
)

func malleateTxs(txConf client.TxConfig, squareSize uint64, txs parsedTxs) (parsedTxs, []*core.Message) {
	var err error
	var msgs []*core.Message
	for i, pTx := range txs {
		if pTx.msg != nil {
			err = pTx.malleate(txConf, squareSize)
			if err != nil {
				txs.remove(i)
				continue
			}
			msgs = append(msgs, pTx.message())
		}
	}
	return txs, msgs
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
