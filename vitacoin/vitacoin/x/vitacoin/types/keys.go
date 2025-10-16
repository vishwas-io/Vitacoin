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