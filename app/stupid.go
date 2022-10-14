package app

import (
	"time"
)

type ShitGenesis struct {
	AppHash  string `json:"app_hash"`
	AppState struct {
		Auth struct {
			Accounts []struct {
				Type        string `json:"@type"`
				BaseAccount struct {
					AccountNumber string      `json:"account_number"`
					Address       string      `json:"address"`
					PubKey        interface{} `json:"pub_key"`
					Sequence      string      `json:"sequence"`
				} `json:"base_account,omitempty"`
				Name          string   `json:"name,omitempty"`
				Permissions   []string `json:"permissions,omitempty"`
				AccountNumber string   `json:"account_number,omitempty"`
				Address       string   `json:"address,omitempty"`
				PubKey        struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"pub_key,omitempty"`
				Sequence string `json:"sequence,omitempty"`
			} `json:"accounts"`
			Params struct {
				MaxMemoCharacters      string `json:"max_memo_characters"`
				SigVerifyCostEd25519   string `json:"sig_verify_cost_ed25519"`
				SigVerifyCostSecp256K1 string `json:"sig_verify_cost_secp256k1"`
				TxSigLimit             string `json:"tx_sig_limit"`
				TxSizeCostPerByte      string `json:"tx_size_cost_per_byte"`
			} `json:"params"`
		} `json:"auth"`
		Bank struct {
			Balances []struct {
				Address string `json:"address"`
				Coins   []struct {
					Amount string `json:"amount"`
					Denom  string `json:"denom"`
				} `json:"coins"`
			} `json:"balances"`
			DenomMetadata []struct {
				Base       string `json:"base"`
				DenomUnits []struct {
					Aliases  []string `json:"aliases"`
					Denom    string   `json:"denom"`
					Exponent int      `json:"exponent"`
				} `json:"denom_units"`
				Description string `json:"description"`
				Display     string `json:"display"`
				Name        string `json:"name"`
				Symbol      string `json:"symbol"`
				URI         string `json:"uri"`
				URIHash     string `json:"uri_hash"`
			} `json:"denom_metadata"`
			Params struct {
				DefaultSendEnabled bool          `json:"default_send_enabled"`
				SendEnabled        []interface{} `json:"send_enabled"`
			} `json:"params"`
			Supply []struct {
				Amount string `json:"amount"`
				Denom  string `json:"denom"`
			} `json:"supply"`
		} `json:"bank"`
		Capability struct {
			Index  string        `json:"index"`
			Owners []interface{} `json:"owners"`
		} `json:"capability"`
		Crisis struct {
			ConstantFee struct {
				Amount string `json:"amount"`
				Denom  string `json:"denom"`
			} `json:"constant_fee"`
		} `json:"crisis"`
		Distribution struct {
			DelegatorStartingInfos []struct {
				DelegatorAddress string `json:"delegator_address"`
				StartingInfo     struct {
					Height         string `json:"height"`
					PreviousPeriod string `json:"previous_period"`
					Stake          string `json:"stake"`
				} `json:"starting_info"`
				ValidatorAddress string `json:"validator_address"`
			} `json:"delegator_starting_infos"`
			DelegatorWithdrawInfos []interface{} `json:"delegator_withdraw_infos"`
			FeePool                struct {
				CommunityPool []struct {
					Amount string `json:"amount"`
					Denom  string `json:"denom"`
				} `json:"community_pool"`
			} `json:"fee_pool"`
			OutstandingRewards []struct {
				OutstandingRewards []interface{} `json:"outstanding_rewards"`
				ValidatorAddress   string        `json:"validator_address"`
			} `json:"outstanding_rewards"`
			Params struct {
				BaseProposerReward  string `json:"base_proposer_reward"`
				BonusProposerReward string `json:"bonus_proposer_reward"`
				CommunityTax        string `json:"community_tax"`
				WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
			} `json:"params"`
			PreviousProposer                string `json:"previous_proposer"`
			ValidatorAccumulatedCommissions []struct {
				Accumulated struct {
					Commission []interface{} `json:"commission"`
				} `json:"accumulated"`
				ValidatorAddress string `json:"validator_address"`
			} `json:"validator_accumulated_commissions"`
			ValidatorCurrentRewards []struct {
				Rewards struct {
					Period  string        `json:"period"`
					Rewards []interface{} `json:"rewards"`
				} `json:"rewards"`
				ValidatorAddress string `json:"validator_address"`
			} `json:"validator_current_rewards"`
			ValidatorHistoricalRewards []struct {
				Period  string `json:"period"`
				Rewards struct {
					CumulativeRewardRatio []interface{} `json:"cumulative_reward_ratio"`
					ReferenceCount        int           `json:"reference_count"`
				} `json:"rewards"`
				ValidatorAddress string `json:"validator_address"`
			} `json:"validator_historical_rewards"`
			ValidatorSlashEvents []interface{} `json:"validator_slash_events"`
		} `json:"distribution"`
		Evidence struct {
			Evidence []interface{} `json:"evidence"`
		} `json:"evidence"`
		Feegrant struct {
			Allowances []interface{} `json:"allowances"`
		} `json:"feegrant"`
		Genutil struct {
			GenTxs []interface{} `json:"gen_txs"`
		} `json:"genutil"`
		Gov struct {
			DepositParams struct {
				MaxDepositPeriod string `json:"max_deposit_period"`
				MinDeposit       []struct {
					Amount string `json:"amount"`
					Denom  string `json:"denom"`
				} `json:"min_deposit"`
			} `json:"deposit_params"`
			Deposits           []interface{} `json:"deposits"`
			Proposals          []interface{} `json:"proposals"`
			StartingProposalID string        `json:"starting_proposal_id"`
			TallyParams        struct {
				Quorum        string `json:"quorum"`
				Threshold     string `json:"threshold"`
				VetoThreshold string `json:"veto_threshold"`
			} `json:"tally_params"`
			Votes        []interface{} `json:"votes"`
			VotingParams struct {
				VotingPeriod string `json:"voting_period"`
			} `json:"voting_params"`
		} `json:"gov"`
		Gqb struct {
		} `json:"gqb"`
		Mint struct {
			Minter struct {
				AnnualProvisions string `json:"annual_provisions"`
				Inflation        string `json:"inflation"`
			} `json:"minter"`
			Params struct {
				BlocksPerYear       string `json:"blocks_per_year"`
				GoalBonded          string `json:"goal_bonded"`
				InflationMax        string `json:"inflation_max"`
				InflationMin        string `json:"inflation_min"`
				InflationRateChange string `json:"inflation_rate_change"`
				MintDenom           string `json:"mint_denom"`
			} `json:"params"`
		} `json:"mint"`
		Params  interface{} `json:"params"`
		Payment struct {
		} `json:"payment"`
		Slashing struct {
			MissedBlocks []struct {
				Address      string        `json:"address"`
				MissedBlocks []interface{} `json:"missed_blocks"`
			} `json:"missed_blocks"`
			Params struct {
				DowntimeJailDuration    string `json:"downtime_jail_duration"`
				MinSignedPerWindow      string `json:"min_signed_per_window"`
				SignedBlocksWindow      string `json:"signed_blocks_window"`
				SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
				SlashFractionDowntime   string `json:"slash_fraction_downtime"`
			} `json:"params"`
			SigningInfos []struct {
				Address              string `json:"address"`
				ValidatorSigningInfo struct {
					Address             string    `json:"address"`
					IndexOffset         string    `json:"index_offset"`
					JailedUntil         time.Time `json:"jailed_until"`
					MissedBlocksCounter string    `json:"missed_blocks_counter"`
					StartHeight         string    `json:"start_height"`
					Tombstoned          bool      `json:"tombstoned"`
				} `json:"validator_signing_info"`
			} `json:"signing_infos"`
		} `json:"slashing"`
		Staking struct {
			Delegations []struct {
				DelegatorAddress string `json:"delegator_address"`
				Shares           string `json:"shares"`
				ValidatorAddress string `json:"validator_address"`
			} `json:"delegations"`
			Exported            bool   `json:"exported"`
			LastTotalPower      string `json:"last_total_power"`
			LastValidatorPowers []struct {
				Address string `json:"address"`
				Power   string `json:"power"`
			} `json:"last_validator_powers"`
			Params struct {
				BondDenom         string `json:"bond_denom"`
				HistoricalEntries int    `json:"historical_entries"`
				MaxEntries        int    `json:"max_entries"`
				MaxValidators     int    `json:"max_validators"`
				MinCommissionRate string `json:"min_commission_rate"`
				UnbondingTime     string `json:"unbonding_time"`
			} `json:"params"`
			Redelegations        []interface{} `json:"redelegations"`
			UnbondingDelegations []interface{} `json:"unbonding_delegations"`
			Validators           []struct {
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
			} `json:"validators"`
		} `json:"staking"`
		Upgrade struct {
		} `json:"upgrade"`
		Vesting struct {
		} `json:"vesting"`
	} `json:"app_state"`
	ChainID         string `json:"chain_id"`
	ConsensusParams struct {
		Block struct {
			MaxBytes   string `json:"max_bytes"`
			MaxGas     string `json:"max_gas"`
			TimeIotaMs int64  `json:"time_iota_ms"`
		} `json:"block"`
		Evidence struct {
			MaxAgeDuration  string `json:"max_age_duration"`
			MaxAgeNumBlocks string `json:"max_age_num_blocks"`
			MaxBytes        string `json:"max_bytes"`
		} `json:"evidence"`
		Validator struct {
			PubKeyTypes []string `json:"pub_key_types"`
		} `json:"validator"`
		Version struct {
		} `json:"version"`
	} `json:"consensus_params"`
	GenesisTime   time.Time `json:"genesis_time"`
	InitialHeight string    `json:"initial_height"`
	Validators    []struct {
		Address string `json:"address"`
		Name    string `json:"name"`
		Power   string `json:"power"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
	} `json:"validators"`
}
