package shares

import (
	"errors"

	"github.com/tendermint/tendermint/pkg/consts"
	coretypes "github.com/tendermint/tendermint/types"
)

var (
	ErrIncorrectNumberOfIndexes = errors.New(
		"number of malleated transations is not identical to the number of wrapped transactions",
	)
	ErrUnexpectedFirstMessageShareIndex = errors.New(
		"the first message started at an unexpected index",
	)
)

func Split(data coretypes.Data) ([][]byte, error) {
	if data.OriginalSquareSize == 0 || !powerOf2(data.OriginalSquareSize) {
		return nil, errors.New("square size is not a power of two")
	}
	wantShareCount := int(data.OriginalSquareSize * data.OriginalSquareSize)
	currentShareCount := 0

	txShares := SplitTxs(data.Txs)
	currentShareCount += len(txShares)

	evdShares, err := SplitEvidence(data.Evidence.Evidence)
	if err != nil {
		return nil, err
	}
	currentShareCount += len(evdShares)

	msgIndexes := ExtractShareIndexes(data.Txs)
	if len(msgIndexes) != len(data.Messages.MessagesList) {
		return nil, ErrIncorrectNumberOfIndexes
	}

	var msgShares [][]byte
	if len(msgIndexes) != 0 {
		if int(msgIndexes[0]) != currentShareCount {
			return nil, ErrUnexpectedFirstMessageShareIndex
		}

		msgShares, err = SplitMessages(msgIndexes, data.Messages.MessagesList)
		if err != nil {
			return nil, err
		}
		currentShareCount += len(msgShares)
	}

	tailShares := TailPaddingShares(wantShareCount - currentShareCount).RawShares()

	// todo: optimize using a predefined slice
	shares := append(append(append(
		txShares,
		evdShares...),
		msgShares...),
		tailShares...)

	return shares, nil
}

func ExtractShareIndexes(txs coretypes.Txs) []uint32 {
	msgIndexes := []uint32{}
	for _, rawTx := range txs {
		if malleatedTx, isMalleated := coretypes.UnwrapMalleatedTx(rawTx); isMalleated {
			msgIndexes = append(msgIndexes, malleatedTx.ShareIndex)
		}
	}

	return msgIndexes
}

func SplitTxs(txs coretypes.Txs) [][]byte {
	writer := NewContiguousShareSplitter(consts.TxNamespaceID)
	for _, tx := range txs {
		writer.WriteTx(tx)
	}
	return writer.Export().RawShares()
}

func SplitEvidence(evd coretypes.EvidenceList) ([][]byte, error) {
	writer := NewContiguousShareSplitter(consts.EvidenceNamespaceID)
	var err error
	for _, ev := range evd {
		err = writer.WriteEvidence(ev)
		if err != nil {
			return nil, err
		}
	}
	return writer.Export().RawShares(), nil
}

func SplitMessages(indexes []uint32, msgs []coretypes.Message) ([][]byte, error) {
	if indexes != nil && len(indexes) != len(msgs) {
		return nil, ErrIncorrectNumberOfIndexes
	}
	writer := NewMessageShareSplitter()
	for i, msg := range msgs {
		writer.Write(msg)
		if indexes != nil && len(indexes) > i+1 {
			writer.WriteNamespacedPaddedShares(int(indexes[i+1]) - writer.Count())
		}
	}
	return writer.Export().RawShares(), nil
}
