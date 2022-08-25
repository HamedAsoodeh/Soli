package shares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	coretypes "github.com/tendermint/tendermint/types"
)

func TestDelimLen(t *testing.T) {
	tests := []uint64{
		1, 2, 3, 4, 5, 6, 7,
		62252,
	}
	for _, tt := range tests {
		res := DelimLen(tt)
		msg := coretypes.Message{
			NamespaceID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Data:        tmrand.Bytes(int(tt)),
		}

		newmsg, err := msg.MarshalDelimited()
		require.NoError(t, err)

		assert.Equal(t, len(newmsg)-int(tt), res)
	}
}
