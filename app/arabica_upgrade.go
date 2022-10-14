package app

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/celestiaorg/celestia-app/app/encoding"
	qgbtypes "github.com/celestiaorg/celestia-app/x/qgb/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/types"
)

func MigrateGenesisStatev070(oldGenPath, newGenPath string) error {
	encCfg := encoding.MakeConfig(ModuleEncodingRegisters...)

	// fix the old genesis by manually adding the timeiota...
	// var b ShitGenesis
	// f, err := os.OpenFile(oldGenPath, os.O_RDWR, os.ModePerm)
	// if err != nil {
	// 	return err
	// }
	// err = json.NewDecoder(f).Decode(&b)
	// if err != nil {
	// 	return err
	// }
	// b.ConsensusParams.Block.TimeIotaMs = 1000
	// err = json.NewEncoder(f).Encode(b)
	// if err != nil {
	// 	return err
	// }
	// f.Close()

	doc, err := types.GenesisDocFromFile(oldGenPath)
	if err != nil {
		return err
	}
	doc.ConsensusParams.Block.TimeIotaMs = 1000

	var appState map[string]json.RawMessage
	err = json.Unmarshal(doc.AppState, &appState)
	if err != nil {
		return err
	}
	appState["qgb"] = qgbGenState(encCfg.Codec)
	appState["staking"] = fillValidatorOrchestratorFields(encCfg.Codec, appState["staking"])

	rawAppState, err := tmjson.Marshal(appState)
	if err != nil {
		return err
	}

	doc.AppState = rawAppState

	encoded, err := tmjson.Marshal(doc)
	if err != nil {
		return err
	}

	// f, err := os.OpenFile(newGenPath, os.O_RDWR, os.ModePerm)
	// if err != nil {
	// 	return err
	// }

	g := string(sdk.MustSortJSON(encoded))
	// defer f.Close()
	fmt.Println(g)

	return nil
}

func qgbGenState(codec codec.Codec) json.RawMessage {
	return codec.MustMarshalJSON(qgbtypes.DefaultGenesis())
}

func fillValidatorOrchestratorFields(codec codec.Codec, msg json.RawMessage) json.RawMessage {
	// var s stakingtypes.GenesisState
	var s map[string]json.RawMessage
	err := json.Unmarshal(msg, &s)
	if err != nil {
		panic(err)
	}
	var vs []Validator
	err = json.Unmarshal(s["validators"], &vs)
	if err != nil {
		panic(err)
	}
	for i, val := range vs {
		val.Orchestrator = val.OperatorAddress
		val.EthAddress = randomEthAddress()
		vs[i] = val
	}
	bzVS, err := json.Marshal(vs)
	if err != nil {
		panic(err)
	}
	s["validators"] = bzVS
	blob, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return blob
}

func randomEthAddress() string {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address
}

// we have to use a custom struct here because otherwise the other roundtrip
// encoding will not work
type Validator struct {
	Commission struct {
		CommissionRates struct {
			MaxChangeRate string `json:"max_change_rate"`
			MaxRate       string `json:"max_rate"`
			Rate          string `json:"rate"`
		} `json:"commission_rates"`
		UpdateTime time.Time `json:"update_time"`
	} `json:"commission"`
	ConsensusPubkey struct {
		Type string `json:"@type"`
		Key  string `json:"key"`
	} `json:"consensus_pubkey"`
	DelegatorShares string `json:"delegator_shares"`
	Description     struct {
		Details         string `json:"details"`
		Identity        string `json:"identity"`
		Moniker         string `json:"moniker"`
		SecurityContact string `json:"security_contact"`
		Website         string `json:"website"`
	} `json:"description"`
	Jailed            bool      `json:"jailed"`
	MinSelfDelegation string    `json:"min_self_delegation"`
	OperatorAddress   string    `json:"operator_address"`
	Status            string    `json:"status"`
	Tokens            string    `json:"tokens"`
	UnbondingHeight   string    `json:"unbonding_height"`
	UnbondingTime     time.Time `json:"unbonding_time"`
	Orchestrator      string    `json:"orchestrator"`
	EthAddress        string    `json:"eth_address"`
}
