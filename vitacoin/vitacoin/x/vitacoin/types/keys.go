package types

const (
	// ModuleName defines the module name
	ModuleName = "vitacoin"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_vitacoin"

	// TreasuryModuleName defines the treasury module account name
	TreasuryModuleName = "vitacoin_treasury"
)

// Store key prefixes
var (
	// ParamsKey is the prefix for params storage
	ParamsKey = []byte{0x01}
	
	// MerchantKeyPrefix is the prefix for merchant storage
	MerchantKeyPrefix = []byte{0x02}
	
	// PaymentKeyPrefix is the prefix for payment storage
	PaymentKeyPrefix = []byte{0x03}
	
	// VaultKeyPrefix is the prefix for vault storage
	VaultKeyPrefix = []byte{0x04}
	
	// RewardPoolKeyPrefix is the prefix for reward pool storage
	RewardPoolKeyPrefix = []byte{0x05}
	
	// Phase 3: Fee & Treasury Keys
	
	// BlockFeeAccumulatorKey stores the current block's fee accumulator
	BlockFeeAccumulatorKey = []byte{0x06}
	
	// FeeStatisticsKey stores cumulative fee statistics
	FeeStatisticsKey = []byte{0x07}
	
	// BurnStatisticsKey stores burn mechanism statistics
	BurnStatisticsKey = []byte{0x08}
	
	// SupplySnapshotPrefix is the prefix for supply snapshot storage (by height)
	SupplySnapshotPrefix = []byte{0x09}
	
	// FeeAccumulatorPrefix is the prefix for historical fee accumulators (by epoch/day)
	FeeAccumulatorPrefix = []byte{0x0A}
	
	// Phase 3 Task 3.4: Treasury Keys
	
	// TreasurySpendingKeyPrefix is the prefix for treasury spending records
	TreasurySpendingKeyPrefix = []byte{0x0B}
)

// GetMerchantKey returns the store key for a specific merchant
func GetMerchantKey(address string) []byte {
	return append(MerchantKeyPrefix, []byte(address)...)
}

// GetPaymentKey returns the store key for a specific payment
func GetPaymentKey(id string) []byte {
	return append(PaymentKeyPrefix, []byte(id)...)
}

// GetVaultKey returns the store key for a specific vault
func GetVaultKey(id string) []byte {
	return append(VaultKeyPrefix, []byte(id)...)
}

// GetRewardPoolKey returns the store key for a specific reward pool
func GetRewardPoolKey(id string) []byte {
	return append(RewardPoolKeyPrefix, []byte(id)...)
}

// Phase 3 Task 3.4: Treasury Key Getters

// GetTreasurySpendingKey returns the store key for a treasury spending record
func GetTreasurySpendingKey(id string) []byte {
	return append(TreasurySpendingKeyPrefix, []byte(id)...)
}