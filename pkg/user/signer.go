package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/celestia-app/app/encoding"
	blob "github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

const defaultPollTime = 3 * time.Second

// Signer is an abstraction for building, signing, and broadcasting Celestia transactions
type Signer struct {
	keys          keyring.Keyring
	address       sdktypes.AccAddress
	enc           client.TxConfig
	grpc          *grpc.ClientConn
	pk            cryptotypes.PubKey
	chainID       string
	accountNumber uint64
	pollTime      time.Duration

	mtx                   sync.RWMutex
	lastSignedSequence    uint64
	lastConfirmedSequence uint64
}

// NewSigner returns a new signer using the provided keyring
func NewSigner(
	keys keyring.Keyring,
	conn *grpc.ClientConn,
	address sdktypes.AccAddress,
	enc client.TxConfig,
	chainID string,
	accountNumber uint64,
	sequence uint64,
) (*Signer, error) {
	// check that the address exists
	record, err := keys.KeyByAddress(address)
	if err != nil {
		return nil, err
	}

	pk, err := record.GetPubKey()
	if err != nil {
		return nil, err
	}

	return &Signer{
		keys:                  keys,
		address:               address,
		grpc:                  conn,
		enc:                   enc,
		pk:                    pk,
		chainID:               chainID,
		accountNumber:         accountNumber,
		lastSignedSequence:    sequence,
		lastConfirmedSequence: sequence,
		pollTime:              defaultPollTime,
	}, nil
}

// SetupSingleSigner sets up a signer based on the provided keyring. The keyring
// must contain exactly one key. It extracts the address from the key and uses
// the grpc connection to populate the chainID, account number, and sequence
// number.
func SetupSingleSigner(ctx context.Context, keys keyring.Keyring, conn *grpc.ClientConn, encCfg encoding.Config) (*Signer, error) {
	records, err := keys.List()
	if err != nil {
		return nil, err
	}

	if len(records) != 1 {
		return nil, errors.New("keyring must contain exactly one key")
	}

	address, err := records[0].GetAddress()
	if err != nil {
		return nil, err
	}

	return SetupSigner(ctx, keys, conn, address, encCfg)
}

// SetupSigner uses the underlying grpc connection to populate the chainID, accountNumber and sequence number of the
// account.
func SetupSigner(
	ctx context.Context,
	keys keyring.Keyring,
	conn *grpc.ClientConn,
	address sdktypes.AccAddress,
	encCfg encoding.Config,
) (*Signer, error) {
	resp, err := tmservice.NewServiceClient(conn).GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return nil, err
	}

	chainID := resp.SdkBlock.Header.ChainID
	accNum, seqNum, err := QueryAccount(ctx, conn, encCfg, address.String())
	if err != nil {
		return nil, err
	}

	return NewSigner(keys, conn, address, encCfg.TxConfig, chainID, accNum, seqNum)
}

func (s *Signer) SubmitTx(ctx context.Context, msgs []sdktypes.Msg, opts ...TxOption) (*sdktypes.TxResponse, error) {
	txBytes, err := s.CreateTx(msgs, opts...)
	if err != nil {
		return nil, err
	}

	resp, err := s.BroadcastTx(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	return s.ConfirmTx(ctx, resp.TxHash)
}

func (s *Signer) SubmitPayForBlob(ctx context.Context, blobs []*tmproto.Blob, opts ...TxOption) (*sdktypes.TxResponse, error) {
	txBytes, err := s.CreatePayForBlob(blobs, opts...)
	if err != nil {
		return nil, err
	}

	resp, err := s.BroadcastTx(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	return s.ConfirmTx(ctx, resp.TxHash)
}

func (s *Signer) CreateTx(msgs []sdktypes.Msg, opts ...TxOption) ([]byte, error) {
	txBuilder := s.txBuilder(opts...)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	if err := s.signTransaction(txBuilder); err != nil {
		return nil, err
	}

	return s.enc.TxEncoder()(txBuilder.GetTx())
}

func (s *Signer) CreatePayForBlob(blobs []*tmproto.Blob, opts ...TxOption) ([]byte, error) {
	msg, err := blob.NewMsgPayForBlobs(s.address.String(), blobs...)
	if err != nil {
		return nil, err
	}

	txBytes, err := s.CreateTx([]sdktypes.Msg{msg}, opts...)
	if err != nil {
		return nil, err
	}

	return tmtypes.MarshalBlobTx(txBytes, blobs...)
}

func (s *Signer) BroadcastTx(ctx context.Context, txBytes []byte) (*sdktypes.TxResponse, error) {
	if s.grpc == nil {
		return nil, errors.New("grpc connection is nil")
	}

	txClient := tx.NewServiceClient(s.grpc)

	// TODO (@cmwaters): handle nonce mismatch errors
	resp, err := txClient.BroadcastTx(
		ctx,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.TxResponse, nil
}

func (s *Signer) ConfirmTx(ctx context.Context, txHash string) (*sdktypes.TxResponse, error) {
	txClient := tx.NewServiceClient(s.grpc)

	resp, err := txClient.GetTx(
		ctx,
		&tx.GetTxRequest{
			Hash: txHash,
		},
	)
	if err == nil {
		return resp.TxResponse, err
	}

	// this is a bit brittle
	if !strings.Contains(err.Error(), "not found") {
		return &sdktypes.TxResponse{}, err
	}

	timer := time.NewTimer(s.pollTime)
	for {
		select {
		case <-ctx.Done():
			return &sdktypes.TxResponse{}, ctx.Err()
		case <-timer.C:
			resp, err = txClient.GetTx(
				ctx,
				&tx.GetTxRequest{
					Hash: txHash,
				},
			)

			if err == nil {
				return resp.TxResponse, err
			}

			if !strings.Contains(err.Error(), "not found") {
				return &sdktypes.TxResponse{}, err
			}
		}
	}
}

func (s *Signer) ChainID() string {
	return s.chainID
}

func (s *Signer) AccountNumber() uint64 {
	return s.accountNumber
}

func (s *Signer) Address() sdktypes.AccAddress {
	return s.address
}

func (s *Signer) signTransaction(builder client.TxBuilder) error {
	signers := builder.GetTx().GetSigners()
	if len(signers) != 1 {
		return fmt.Errorf("expected 1 signer, got %d", len(signers))
	}

	if !s.address.Equals(signers[0]) {
		return fmt.Errorf("expected signer %s, got %s", s.address.String(), signers[0].String())
	}

	sequence := s.getSequence()

	// To ensure we have the correct bytes to sign over we produce
	// a dry run of the signing data
	draftsigV2 := signing.SignatureV2{
		PubKey: s.pk,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: sequence,
	}

	err := builder.SetSignatures(draftsigV2)
	if err != nil {
		return fmt.Errorf("error setting draft signatures: %w", err)
	}

	// now we can use the data to produce the signature from each signer
	signature, err := s.createSignature(builder, sequence)
	if err != nil {
		return fmt.Errorf("error creating signature: %w", err)
	}
	sigV2 := signing.SignatureV2{
		PubKey: s.pk,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signature,
		},
		Sequence: sequence,
	}

	err = builder.SetSignatures(sigV2)
	if err != nil {
		return fmt.Errorf("error setting signatures: %w", err)
	}

	return nil
}

func (s *Signer) createSignature(builder client.TxBuilder, sequence uint64) ([]byte, error) {
	signerData := authsigning.SignerData{
		Address:       s.address.String(),
		ChainID:       s.ChainID(),
		AccountNumber: s.accountNumber,
		Sequence:      sequence,
		PubKey:        s.pk,
	}

	bytesToSign, err := s.enc.SignModeHandler().GetSignBytes(
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		builder.GetTx(),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting sign bytes: %w", err)
	}

	signature, _, err := s.keys.SignByAddress(s.address, bytesToSign)
	if err != nil {
		return nil, fmt.Errorf("error signing bytes: %w", err)
	}

	return signature, nil
}

// NewTxBuilder returns the default sdk Tx builder using the celestia-app encoding config
func (s *Signer) txBuilder(opts ...TxOption) client.TxBuilder {
	builder := s.enc.NewTxBuilder()
	for _, opt := range opts {
		builder = opt(builder)
	}
	return builder
}

func (s *Signer) getSequence() uint64 {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	defer func() { s.lastSignedSequence++ }()
	return s.lastSignedSequence
}

// QueryAccount fetches the account number and sequence number from the celestia-app node.
func QueryAccount(ctx context.Context, conn *grpc.ClientConn, encCfg encoding.Config, address string) (accNum uint64, seqNum uint64, err error) {
	qclient := authtypes.NewQueryClient(conn)
	resp, err := qclient.Account(
		ctx,
		&authtypes.QueryAccountRequest{Address: address},
	)
	if err != nil {
		return accNum, seqNum, err
	}

	var acc authtypes.AccountI
	err = encCfg.InterfaceRegistry.UnpackAny(resp.Account, &acc)
	if err != nil {
		return accNum, seqNum, err
	}

	accNum, seqNum = acc.GetAccountNumber(), acc.GetSequence()
	return accNum, seqNum, nil
}
