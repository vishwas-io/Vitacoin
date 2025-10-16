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
	AttributeKeyTreasuryAmount  = "treasury_amount"

	// General attributes
	AttributeKeyHeight          = "height"
	AttributeKeyBlockTime       = "block_time"
	AttributeKeyModule          = "module"
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