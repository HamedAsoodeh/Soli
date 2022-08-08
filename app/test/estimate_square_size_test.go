package app_test

// func TestEstimateSquareSize(t *testing.T) {
// 	type test struct {
// 		name                  string
// 		wPFDCount, messgeSize int
// 		expectedSize          uint64
// 	}
// 	tests := []test{
// 		{"empty block minimum square size", 0, 0, consts.MinSquareSize},
// 		{"random small block square size 2", 1, 400, 2},
// 		{"random small block square size 4", 1, 2000, 4},
// 		{"random small block square size 4", 4, 2000, 8},
// 		{"random medium block square size 32", 50, 2000, 32},
// 		{"full block max square size", 16000, 200, consts.MaxSquareSize},
// 		{"overly full block", 16000, 1000, consts.MaxSquareSize},
// 	}
// 	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
// 	signer := testutil.GenerateKeyringSigner(t, "estimate-key")
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			txs := generateManyRawWirePFD(t, encConf.TxConfig, signer, tt.wPFDCount, tt.messgeSize)
// 			res := app.EstimateSquareSize(&core.Data{Txs: txs}, encConf.TxConfig)
// 			assert.Equal(t, tt.expectedSize, res)
// 		})
// 	}

// }

// func generateManyRawWirePFD(t *testing.T, txConf client.TxConfig, signer *types.KeyringSigner, count, size int) [][]byte {
// 	txs := make([][]byte, count)
// 	for i := 0; i < count; i++ {
// 		msg := tmrand.Bytes(size)
// 		ns := randomValidNamespace()
// 		txs[i] = testtxs.GenerateRawWirePFDTx(t, txConf, ns, msg, signer)
// 	}

// 	return txs
// }
