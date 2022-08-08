package app

import (
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// parsedTx is an interanl struct that keeps track of potentially valid txs and
// their wire messages if they have any.
type parsedTx struct {
	// the original raw bytes of the tx
	rawTx []byte
	// tx is the parsed sdk tx. this is nil for all txs that do not contain a
	// MsgWirePayForData
	tx signing.Tx
	// msg is the wire msg if it exists in the tx. This field is nil for all txs
	// that do not contain one.
	msg *types.MsgWirePayForData
}

// parseTxs decodes raw tendermint txs along with checking if they contain any
// MsgWirePayForData txs. If a MsgWirePayForData is found in the tx, then it is
// saved in the parsedTx that is returned. It ignores invalid txs completely.
func parseTxs(conf client.TxConfig, rawTxs [][]byte) []parsedTx {
	parsedTxs := []parsedTx{}
	for _, rawTx := range rawTxs {
		tx, err := conf.TxDecoder()(rawTx)
		if err != nil {
			continue
		}

		authTx, ok := tx.(signing.Tx)
		if !ok {
			continue
		}

		pTx := parsedTx{
			rawTx: rawTx,
		}

		wireMsg, err := types.ExtractMsgWirePayForData(authTx)
		if err != nil {
			// we catch this error because it means that there are no
			// potentially valid MsgWirePayForData messages in this tx. We still
			// want to keep this tx, so we append it to the parsed txs.
			parsedTxs = append(parsedTxs, pTx)
			continue
		}

		// run basic validation on the message
		err = wireMsg.ValidateBasic()
		if err != nil {
			continue
		}

		// run basic validation on the transaction
		err = authTx.ValidateBasic()
		if err != nil {
			continue
		}

		pTx.tx = authTx
		pTx.msg = wireMsg
		parsedTxs = append(parsedTxs, pTx)
	}
	return parsedTxs
}
