package types

// Event types for the vitacoin module
const (
	// Legacy transaction events (deprecated)
	EventTypeCreateTransaction = "create_transaction"
	EventTypeUpdateTransaction = "update_transaction"
	EventTypeDeleteTransaction = "delete_transaction"

	// Payment events
	EventTypePaymentCreated   = "payment_created"
	EventTypePaymentCompleted = "payment_completed"
	EventTypePaymentRefunded  = "payment_refunded"
	EventTypePaymentExpired   = "payment_expired"
	EventTypePaymentSettled   = "payment_settled"

	// Merchant events
	EventTypeMerchantRegistered = "merchant_registered"
	EventTypeMerchantUpdated    = "merchant_updated"
	EventTypeMerchantTierChanged = "merchant_tier_changed"

	// Vault events
	EventTypeVaultCreated   = "vault_created"
	EventTypeVaultWithdrawn = "vault_withdrawn"
	EventTypeVaultUnlocked  = "vault_unlocked"

	// Reward pool events
	EventTypeRewardPoolCreated     = "reward_pool_created"
	EventTypeRewardDistributed     = "reward_distributed"
	EventTypeRewardPoolStatusChanged = "reward_pool_status_changed"

	// Fee events
	EventTypeFeeDistributed = "fee_distributed"
	EventTypeFeesBurned     = "fees_burned"

	// Module events
	EventTypeParamsUpdated = "params_updated"
	
	// Phase 3: Fee & Treasury Events
	EventTypeFeeDistribution     = "fee_distribution"
	EventTypeFeeBurned           = "fee_burned"
	EventTypeTreasuryDeposit     = "treasury_deposit"
	EventTypeSupplySnapshot      = "supply_snapshot"
	EventTypeBurnCapReached      = "burn_cap_reached"
	EventTypeFeeCollectionPaused = "fee_collection_paused"
	
	// Phase 3 Task 3.4: Treasury Events
	EventTypeTreasurySpent       = "treasury_spent"
	EventTypeTreasuryProposal    = "treasury_proposal"

	// Phase 4: Staking Events
	EventTypeDelegation        = "delegation"
	EventTypeUnbonding         = "unbonding"
	EventTypeUnbondingReleased = "unbonding_released"

	// Phase 4: Validator Events
	EventTypeValidatorRegistered = "validator_registered"
	EventTypeValidatorSlashed    = "validator_slashed"
	EventTypeValidatorJailed     = "validator_jailed"
	EventTypeValidatorUnjailed   = "validator_unjailed"

	// Phase 4: Reward Events
	EventTypeRewardClaim        = "reward_claim"
	EventTypeRewardDistribution = "reward_distribution"

	// Phase 4: Liquid Staking Events
	EventTypeLiquidDelegate   = "liquid_delegate"
	EventTypeLiquidUndelegate = "liquid_undelegate"
)

// Attribute keys for events
const (
	// Legacy attributes (deprecated)
	AttributeKeyTransactionID = "transaction_id"
	AttributeKeySender        = "sender"
	AttributeKeyReceiver      = "receiver"
	AttributeKeyAmount        = "amount"

	// Payment attributes
	AttributeKeyPaymentId     = "payment_id"
	AttributeKeyFromAddress   = "from_address"
	AttributeKeyToAddress     = "to_address"
	AttributeKeyPaymentAmount = "payment_amount"
	AttributeKeyPaymentFee    = "payment_fee"
	AttributeKeyPaymentStatus = "payment_status"
	AttributeKeyPaymentMemo   = "payment_memo"

	// Merchant attributes
	AttributeKeyMerchantAddress   = "merchant_address"
	AttributeKeyBusinessName      = "business_name"
	AttributeKeyMerchantTier      = "merchant_tier"
	AttributeKeyStakeAmount       = "stake_amount"
	AttributeKeyTotalVolume       = "total_volume"
	AttributeKeyOldTier          = "old_tier"
	AttributeKeyNewTier          = "new_tier"

	// Vault attributes
	AttributeKeyVaultId        = "vault_id"
	AttributeKeyOwner          = "owner"
	AttributeKeyVaultAmount    = "vault_amount"
	AttributeKeyLockDuration   = "lock_duration"
	AttributeKeyUnlockHeight   = "unlock_height"
	AttributeKeyRewardMultiplier = "reward_multiplier"

	// Reward pool attributes
	AttributeKeyPoolId          = "pool_id"
	AttributeKeyPoolName        = "pool_name"
	AttributeKeyTotalRewards    = "total_rewards"
	AttributeKeyDistributedRewards = "distributed_rewards"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyRewardAmount    = "reward_amount"
	AttributeKeyStartHeight     = "start_height"
	AttributeKeyEndHeight       = "end_height"

	// Fee attributes
	AttributeKeyFeeAmount       = "fee_amount"
	AttributeKeyValidatorShare  = "validator_share"
	AttributeKeyBurnedAmount    = "burned_amount"

	// General attributes
	AttributeKeyHeight          = "height"
	AttributeKeyBlockTime       = "block_time"
	AttributeKeyModule          = "module"
	
	// Phase 3: Payment Settlement Attributes
	AttributeKeyPaymentID      = "payment_id"
	AttributeKeyPayer          = "payer"
	AttributeKeyMerchant       = "merchant"
	AttributeKeyGrossAmount    = "gross_amount"
	AttributeKeyProtocolFee    = "protocol_fee"
	AttributeKeyNetAmount      = "net_amount"
	
	// Phase 3: Fee Distribution Attributes
	AttributeKeyTotalFees       = "total_fees"
	AttributeKeyBurnAmount      = "burn_amount"
	AttributeKeyValidatorAmount = "validator_amount"
	AttributeKeyTreasuryAmount  = "treasury_amount"
	AttributeKeyTransactionCount = "transaction_count"
	
	// Phase 3: Supply Tracking Attributes
	AttributeKeyTotalSupply       = "total_supply"
	AttributeKeyCirculatingSupply = "circulating_supply"
	AttributeKeyLiquidSupply      = "liquid_supply"
	AttributeKeyBondedSupply      = "bonded_supply"
	AttributeKeyBurnedCumulative  = "burned_cumulative"
	
	// Phase 3 Task 3.4: Treasury Attributes
	AttributeKeyProposalId        = "proposal_id"
	AttributeKeyPurpose           = "purpose"
	AttributeKeyTreasuryBalance   = "treasury_balance"
	AttributeKeySpendingId        = "spending_id"

	// Phase 4: Staking Attributes
	AttributeKeyDelegator        = "delegator"
	AttributeKeyValidator        = "validator"
	AttributeKeyDelegationAmount = "delegation_amount"
	AttributeKeyMaturityBlock    = "maturity_block"
	AttributeKeyStartBlock       = "start_block"

	// Phase 4: Validator Attributes
	AttributeKeyOperatorAddress = "operator_address"
	AttributeKeyMoniker         = "moniker"
	AttributeKeyCommission      = "commission"
	AttributeKeySelfBond        = "self_bond"
	AttributeKeySlashFactor     = "slash_factor"
	AttributeKeySlashAmount     = "slash_amount"
	AttributeKeyCreatedBlock    = "created_block"

	// Phase 5: Governance Event Types
	EventTypeProposalSubmitted = "proposal_submitted"
	EventTypeProposalActivated = "proposal_activated"
	EventTypeDepositAdded      = "deposit_added"

	// Phase 5: Governance Attributes
	AttributeKeyProposer       = "proposer"
	AttributeKeyProposalType   = "proposal_type"
	AttributeKeyProposalStatus = "proposal_status"
	AttributeKeyDepositor      = "depositor"
	AttributeKeyDepositAmount  = "deposit_amount"
)

// CreateTransactionEvent is emitted when a new transaction is created.
type CreateTransactionEvent struct {
	TransactionID string
	Sender        string
	Receiver      string
	Amount        string
}

// UpdateTransactionEvent is emitted when a transaction is updated.
type UpdateTransactionEvent struct {
	TransactionID string
	Sender        string
	Receiver      string
	Amount        string
}

// DeleteTransactionEvent is emitted when a transaction is deleted.
type DeleteTransactionEvent struct {
	TransactionID string
}
const (
	// Phase 5: Governance — voting, tally, execution events
	EventTypeVoteCast         = "vote_cast"
	EventTypeProposalTallied  = "proposal_tallied"
	EventTypeProposalExecuted = "proposal_executed"

	// Phase 5: Governance — vote / tally attributes
	AttributeKeyVoter          = "voter"
	AttributeKeyVoteOption     = "vote_option"
	AttributeKeyVoteWeight     = "vote_weight"
	AttributeKeyTallyPassed    = "passed"
	AttributeKeyTallyReason    = "tally_reason"
	AttributeKeyExecutionError = "execution_error"
)
