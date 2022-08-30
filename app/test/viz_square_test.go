package app_test

import (
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/testutil"
	"github.com/stretchr/testify/require"
)

func TestVizualizeSquare(t *testing.T) {
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	signer := testutil.GenerateKeyringSigner(t, "msg-inclusion-key")
	data, err := app.GenerateValidBlockData(t, encConf.TxConfig, signer, 6, 10, 10000)
	require.NoError(t, err)
	rawShares, err := shares.Split(data)
	require.NoError(t, err)
	err = shares.VisualizeSquare("/tmp/square.png", 480, rawShares)
	require.NoError(t, err)
}
