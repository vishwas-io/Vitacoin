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

	// BondDenom is the staking bond denomination for VitaCoin
	BondDenom = "avita"

	// Phase 4: Staking Store Key Prefixes (as byte constants)
	StakingDelegationPrefixByte = 0x10
	StakingUnbondingPrefixByte  = 0x11
	StakingValidatorPrefixByte  = 0x12
	StakingPendingRewardPrefixByte = 0x21

	// Phase 4: Liquid Staking Key Prefix
	StakingStVITASupplyPrefixByte = 0x22
)

// Phase 4: Staking store keys
var (
	// DelegationKeyPrefix is the prefix for delegation records
	DelegationKeyPrefix = []byte{StakingDelegationPrefixByte}

	// UnbondingKeyPrefix is the prefix for unbonding queue entries
	UnbondingKeyPrefix = []byte{StakingUnbondingPrefixByte}

	// ValidatorKeyPrefix is the prefix for validator records
	ValidatorKeyPrefix = []byte{StakingValidatorPrefixByte}

	// PendingRewardKeyPrefix is the prefix for pending delegator reward records
	PendingRewardKeyPrefix = []byte{StakingPendingRewardPrefixByte}

	// StVITASupplyKey is the single KV key for the total stVITA supply
	StVITASupplyKey = []byte{StakingStVITASupplyPrefixByte}
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

	// RateLimitKeyPrefix is the prefix for per-address rate limit tracking (last tx block height)
	RateLimitKeyPrefix = []byte{0x0C}

	// RateLimitConfigKey stores the MinBlocksBetweenTx config value
	RateLimitConfigKey = []byte{0x0D}
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

// GetRateLimitKey returns the store key for a per-address last-tx-block record
func GetRateLimitKey(address string) []byte {
	return append(RateLimitKeyPrefix, []byte(address)...)
}

// Phase 4: Staking Key Getters

// GetValidatorKey returns the store key for a validator record.
func GetValidatorKey(addr string) []byte {
	return append(ValidatorKeyPrefix, []byte(addr)...)
}

// GetDelegationKey returns the store key for a delegation record.
// Key format: DelegationKeyPrefix | delegatorAddr | validatorAddr
func GetDelegationKey(delegatorAddr, validatorAddr string) []byte {
	key := append(DelegationKeyPrefix, []byte(delegatorAddr)...)
	key = append(key, []byte(":")...)
	return append(key, []byte(validatorAddr)...)
}

// GetUnbondingKey returns the store key for an unbonding queue entry.
// Key format: UnbondingKeyPrefix | delegatorAddr | validatorAddr | maturityBlock (8 bytes big-endian)
func GetUnbondingKey(delegatorAddr, validatorAddr string, maturityBlock int64) []byte {
	key := append(UnbondingKeyPrefix, []byte(delegatorAddr)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(validatorAddr)...)
	key = append(key, []byte(":")...)
	// Encode maturityBlock as 8-byte big-endian for lexicographic ordering
	blockBytes := make([]byte, 8)
	blockBytes[0] = byte(maturityBlock >> 56)
	blockBytes[1] = byte(maturityBlock >> 48)
	blockBytes[2] = byte(maturityBlock >> 40)
	blockBytes[3] = byte(maturityBlock >> 32)
	blockBytes[4] = byte(maturityBlock >> 24)
	blockBytes[5] = byte(maturityBlock >> 16)
	blockBytes[6] = byte(maturityBlock >> 8)
	blockBytes[7] = byte(maturityBlock)
	return append(key, blockBytes...)
}

// GetPendingRewardKey returns the store key for a delegator's pending reward record.
func GetPendingRewardKey(delegatorAddr string) []byte {
	return append(PendingRewardKeyPrefix, []byte(delegatorAddr)...)
}

// GetStVITASupplyKey returns the single KV key for total stVITA supply.
func GetStVITASupplyKey() []byte {
	return StVITASupplyKey
}

// ─── Phase 5: Governance Key Prefixes ─────────────────────────────────────────

var (
	// KeyPrefixProposal is the prefix for proposal records (0x30).
	KeyPrefixProposal = []byte{0x30}

	// KeyPrefixVote is the prefix for vote records (0x31).
	KeyPrefixVote = []byte{0x31}

	// KeyPrefixDeposit is the prefix for deposit records (0x32).
	KeyPrefixDeposit = []byte{0x32}

	// KeyProposalCounter is the key for the proposal sequence counter (0x33).
	KeyProposalCounter = []byte{0x33}
)

// GetProposalKey returns the KV store key for a proposal.
// Key layout: KeyPrefixProposal | proposalId (8-byte big-endian)
func GetProposalKey(proposalId uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalId >> 56)
	bz[1] = byte(proposalId >> 48)
	bz[2] = byte(proposalId >> 40)
	bz[3] = byte(proposalId >> 32)
	bz[4] = byte(proposalId >> 24)
	bz[5] = byte(proposalId >> 16)
	bz[6] = byte(proposalId >> 8)
	bz[7] = byte(proposalId)
	return append(KeyPrefixProposal, bz...)
}

// GetVoteKey returns the KV store key for a single vote.
// Key layout: KeyPrefixVote | proposalId (8-byte big-endian) | ":" | voter
func GetVoteKey(proposalId uint64, voter string) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalId >> 56)
	bz[1] = byte(proposalId >> 48)
	bz[2] = byte(proposalId >> 40)
	bz[3] = byte(proposalId >> 32)
	bz[4] = byte(proposalId >> 24)
	bz[5] = byte(proposalId >> 16)
	bz[6] = byte(proposalId >> 8)
	bz[7] = byte(proposalId)
	key := append(KeyPrefixVote, bz...)
	key = append(key, byte(':'))
	return append(key, []byte(voter)...)
}

// GetVotesByProposalPrefix returns the prefix used to iterate all votes for a proposal.
func GetVotesByProposalPrefix(proposalId uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalId >> 56)
	bz[1] = byte(proposalId >> 48)
	bz[2] = byte(proposalId >> 40)
	bz[3] = byte(proposalId >> 32)
	bz[4] = byte(proposalId >> 24)
	bz[5] = byte(proposalId >> 16)
	bz[6] = byte(proposalId >> 8)
	bz[7] = byte(proposalId)
	return append(KeyPrefixVote, bz...)}

// ─── Phase 6: IBC Key Prefixes ────────────────────────────────────────────────

var (
	// KeyPrefixIBCPacket is the prefix for pending outgoing IBC packet records (0x40).
	KeyPrefixIBCPacket = []byte{0x40}
)

// GetIBCPacketKey returns the KV store key for a pending IBC packet.
// Key layout: KeyPrefixIBCPacket | sequence (8-byte big-endian)
func GetIBCPacketKey(sequence uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(sequence >> 56)
	bz[1] = byte(sequence >> 48)
	bz[2] = byte(sequence >> 40)
	bz[3] = byte(sequence >> 32)
	bz[4] = byte(sequence >> 24)
	bz[5] = byte(sequence >> 16)
	bz[6] = byte(sequence >> 8)
	bz[7] = byte(sequence)
	return append(KeyPrefixIBCPacket, bz...)
}
