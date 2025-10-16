package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	
	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

// Keeper of the vitacoin store
// Manages all state transitions and business logic for the VITACOIN module
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	logger       log.Logger
	
	// the address capable of executing a MsgUpdateParams message (typically the gov module)
	authority string
}

// NewKeeper creates a new vitacoin Keeper instance with production-level validation
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
) Keeper {
	// Ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %v", err))
	}

	// Validate that codec is not nil
	if cdc == nil {
		panic("codec cannot be nil")
	}

	// Validate that storeService is not nil
	if storeService == nil {
		panic("store service cannot be nil")
	}

	keeper := Keeper{
		storeService: storeService,
		cdc:          cdc,
		logger:       logger.With("module", types.ModuleName),
		authority:    authority,
	}

	return keeper
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger with context
func (k Keeper) Logger() log.Logger {
	return k.logger
}

// GetStoreService returns the store service for direct access if needed
func (k Keeper) GetStoreService() store.KVStoreService {
	return k.storeService
}

// GetCodec returns the codec for encoding/decoding operations
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// InitGenesis initializes the module's state from a genesis state with full validation
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) error {
	// Validate genesis state
	if err := genState.Validate(); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}
	
	// Set params with validation
	if err := k.SetParams(ctx, genState.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Initialize merchants
	for _, merchant := range genState.MerchantList {
		if err := k.SetMerchant(ctx, merchant); err != nil {
			return fmt.Errorf("failed to set merchant %s: %w", merchant.Address, err)
		}
	}

	// Initialize payments
	for _, payment := range genState.PaymentList {
		if err := k.SetPayment(ctx, payment); err != nil {
			return fmt.Errorf("failed to set payment %s: %w", payment.Id, err)
		}
	}

	// Initialize vaults
	for _, vault := range genState.VaultList {
		if err := k.SetVault(ctx, vault); err != nil {
			return fmt.Errorf("failed to set vault %s: %w", vault.Id, err)
		}
	}

	// Initialize reward pools
	for _, pool := range genState.PoolList {
		if err := k.SetRewardPool(ctx, pool); err != nil {
			return fmt.Errorf("failed to set reward pool %s: %w", pool.Id, err)
		}
	}

	k.logger.Info("initialized vitacoin module from genesis",
		"merchants", len(genState.MerchantList),
		"payments", len(genState.PaymentList),
		"vaults", len(genState.VaultList),
		"pools", len(genState.PoolList),
	)

	return nil
}

// ExportGenesis exports the module's state to a genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}
	
	// Export merchants
	merchants, err := k.GetAllMerchants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all merchants: %w", err)
	}

	// Export payments
	payments, err := k.GetAllPayments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all payments: %w", err)
	}

	// Export vaults
	vaults, err := k.GetAllVaults(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all vaults: %w", err)
	}

	// Export reward pools
	pools, err := k.GetAllRewardPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all reward pools: %w", err)
	}

	k.logger.Info("exported vitacoin module genesis state",
		"merchants", len(merchants),
		"payments", len(payments),
		"vaults", len(vaults),
		"pools", len(pools),
	)

	return &types.GenesisState{
		Params:       params,
		MerchantList: merchants,
		PaymentList:  payments,
		VaultList:    vaults,
		PoolList:     pools,
	}, nil
}

// ValidateAuthority checks if the provided address matches the module authority
func (k Keeper) ValidateAuthority(address string) error {
	if address != k.authority {
		return fmt.Errorf("unauthorized: expected %s, got %s", k.authority, address)
	}
	return nil
}

// --- Merchant Management Methods ---

// SetMerchant stores a merchant in the state with validation
func (k Keeper) SetMerchant(ctx context.Context, merchant types.Merchant) error {
	// Validate merchant address
	if _, err := sdk.AccAddressFromBech32(merchant.Address); err != nil {
		return fmt.Errorf("invalid merchant address: %w", err)
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&merchant)
	
	return store.Set(types.GetMerchantKey(merchant.Address), bz)
}

// GetMerchant retrieves a merchant by address
func (k Keeper) GetMerchant(ctx context.Context, address string) (types.Merchant, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.GetMerchantKey(address))
	if err != nil {
		return types.Merchant{}, err
	}
	
	if bz == nil {
		return types.Merchant{}, fmt.Errorf("merchant not found: %s", address)
	}
	
	var merchant types.Merchant
	k.cdc.MustUnmarshal(bz, &merchant)
	return merchant, nil
}

// HasMerchant checks if a merchant exists
func (k Keeper) HasMerchant(ctx context.Context, address string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	has, err := store.Has(types.GetMerchantKey(address))
	if err != nil {
		return false, err
	}
	
	return has, nil
}

// DeleteMerchant removes a merchant from the state
func (k Keeper) DeleteMerchant(ctx context.Context, address string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetMerchantKey(address))
}

// GetAllMerchants retrieves all merchants from the state
func (k Keeper) GetAllMerchants(ctx context.Context) ([]types.Merchant, error) {
	store := k.storeService.OpenKVStore(ctx)
	merchants := []types.Merchant{}
	
	iter, err := store.Iterator(types.MerchantKeyPrefix, storetypes.PrefixEndBytes(types.MerchantKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	
	for ; iter.Valid(); iter.Next() {
		var merchant types.Merchant
		k.cdc.MustUnmarshal(iter.Value(), &merchant)
		merchants = append(merchants, merchant)
	}
	
	return merchants, nil
}

// --- Payment Management Methods ---

// SetPayment stores a payment in the state
func (k Keeper) SetPayment(ctx context.Context, payment types.Payment) error {
	if payment.Id == "" {
		return fmt.Errorf("payment ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&payment)
	
	return store.Set(types.GetPaymentKey(payment.Id), bz)
}

// GetPayment retrieves a payment by ID
func (k Keeper) GetPayment(ctx context.Context, id string) (types.Payment, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.GetPaymentKey(id))
	if err != nil {
		return types.Payment{}, err
	}
	
	if bz == nil {
		return types.Payment{}, fmt.Errorf("payment not found: %s", id)
	}
	
	var payment types.Payment
	k.cdc.MustUnmarshal(bz, &payment)
	return payment, nil
}

// HasPayment checks if a payment exists
func (k Keeper) HasPayment(ctx context.Context, id string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	has, err := store.Has(types.GetPaymentKey(id))
	if err != nil {
		return false, err
	}
	
	return has, nil
}

// DeletePayment removes a payment from the state
func (k Keeper) DeletePayment(ctx context.Context, id string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetPaymentKey(id))
}

// GetAllPayments retrieves all payments from the state
func (k Keeper) GetAllPayments(ctx context.Context) ([]types.Payment, error) {
	store := k.storeService.OpenKVStore(ctx)
	payments := []types.Payment{}
	
	iter, err := store.Iterator(types.PaymentKeyPrefix, storetypes.PrefixEndBytes(types.PaymentKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	
	for ; iter.Valid(); iter.Next() {
		var payment types.Payment
		k.cdc.MustUnmarshal(iter.Value(), &payment)
		payments = append(payments, payment)
	}
	
	return payments, nil
}

// --- Vault Management Methods ---

// SetVault stores a vault in the state
func (k Keeper) SetVault(ctx context.Context, vault types.Vault) error {
	if vault.Id == "" {
		return fmt.Errorf("vault ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&vault)
	
	return store.Set(types.GetVaultKey(vault.Id), bz)
}

// GetVault retrieves a vault by ID
func (k Keeper) GetVault(ctx context.Context, id string) (types.Vault, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.GetVaultKey(id))
	if err != nil {
		return types.Vault{}, err
	}
	
	if bz == nil {
		return types.Vault{}, fmt.Errorf("vault not found: %s", id)
	}
	
	var vault types.Vault
	k.cdc.MustUnmarshal(bz, &vault)
	return vault, nil
}

// HasVault checks if a vault exists
func (k Keeper) HasVault(ctx context.Context, id string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	has, err := store.Has(types.GetVaultKey(id))
	if err != nil {
		return false, err
	}
	
	return has, nil
}

// DeleteVault removes a vault from the state
func (k Keeper) DeleteVault(ctx context.Context, id string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetVaultKey(id))
}

// GetAllVaults retrieves all vaults from the state
func (k Keeper) GetAllVaults(ctx context.Context) ([]types.Vault, error) {
	store := k.storeService.OpenKVStore(ctx)
	vaults := []types.Vault{}
	
	iter, err := store.Iterator(types.VaultKeyPrefix, storetypes.PrefixEndBytes(types.VaultKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	
	for ; iter.Valid(); iter.Next() {
		var vault types.Vault
		k.cdc.MustUnmarshal(iter.Value(), &vault)
		vaults = append(vaults, vault)
	}
	
	return vaults, nil
}

// --- Reward Pool Management Methods ---

// SetRewardPool stores a reward pool in the state
func (k Keeper) SetRewardPool(ctx context.Context, pool types.RewardPool) error {
	if pool.Id == "" {
		return fmt.Errorf("reward pool ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&pool)
	
	return store.Set(types.GetRewardPoolKey(pool.Id), bz)
}

// GetRewardPool retrieves a reward pool by ID
func (k Keeper) GetRewardPool(ctx context.Context, id string) (types.RewardPool, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.GetRewardPoolKey(id))
	if err != nil {
		return types.RewardPool{}, err
	}
	
	if bz == nil {
		return types.RewardPool{}, fmt.Errorf("reward pool not found: %s", id)
	}
	
	var pool types.RewardPool
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, nil
}

// HasRewardPool checks if a reward pool exists
func (k Keeper) HasRewardPool(ctx context.Context, id string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	has, err := store.Has(types.GetRewardPoolKey(id))
	if err != nil {
		return false, err
	}
	
	return has, nil
}

// DeleteRewardPool removes a reward pool from the state
func (k Keeper) DeleteRewardPool(ctx context.Context, id string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetRewardPoolKey(id))
}

// GetAllRewardPools retrieves all reward pools from the state
func (k Keeper) GetAllRewardPools(ctx context.Context) ([]types.RewardPool, error) {
	store := k.storeService.OpenKVStore(ctx)
	pools := []types.RewardPool{}
	
	iter, err := store.Iterator(types.RewardPoolKeyPrefix, storetypes.PrefixEndBytes(types.RewardPoolKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	
	for ; iter.Valid(); iter.Next() {
		var pool types.RewardPool
		k.cdc.MustUnmarshal(iter.Value(), &pool)
		pools = append(pools, pool)
	}
	
	return pools, nil
}

// BeginBlocker is called at the beginning of each block
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Log block begin for vitacoin module
	k.logger.Debug("BeginBlock", "height", ctx.BlockHeight())
	
	// Process any time-based operations
	if err := k.processTimeBasedOperations(ctx); err != nil {
		k.logger.Error("failed to process time-based operations", "error", err)
		return fmt.Errorf("failed to process time-based operations: %w", err)
	}
	
	return nil
}

// EndBlocker is called at the end of each block
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// Log block end for vitacoin module
	k.logger.Debug("EndBlock", "height", ctx.BlockHeight())
	
	// Process fee distribution and rewards
	if err := k.processFeeDistribution(ctx); err != nil {
		k.logger.Error("failed to process fee distribution", "error", err)
		return fmt.Errorf("failed to process fee distribution: %w", err)
	}
	
	// Update merchant tiers based on volume
	if err := k.updateMerchantTiers(ctx); err != nil {
		k.logger.Error("failed to update merchant tiers", "error", err)
		return fmt.Errorf("failed to update merchant tiers: %w", err)
	}
	
	// Expire old payments
	if err := k.expireOldPayments(ctx); err != nil {
		k.logger.Error("failed to expire old payments", "error", err)
		return fmt.Errorf("failed to expire old payments: %w", err)
	}
	
	return nil
}

// processTimeBasedOperations handles operations that need to be executed based on time/height
func (k Keeper) processTimeBasedOperations(ctx sdk.Context) error {
	// Process vault unlocks
	if err := k.processVaultUnlocks(ctx); err != nil {
		return fmt.Errorf("failed to process vault unlocks: %w", err)
	}
	
	// Process reward pool activations/deactivations
	if err := k.processRewardPoolStatus(ctx); err != nil {
		return fmt.Errorf("failed to process reward pool status: %w", err)
	}
	
	return nil
}

// processFeeDistribution distributes collected fees according to the module parameters
func (k Keeper) processFeeDistribution(ctx sdk.Context) error {
	// Implementation will be added in later tasks
	// This is where we'll implement the 50/25/25 fee split:
	// - 50% to validators (staking module)
	// - 25% to burn address
	// - 25% to treasury
	
	k.logger.Debug("Processing fee distribution", "height", ctx.BlockHeight())
	return nil
}

// updateMerchantTiers updates merchant tiers based on their transaction volume
func (k Keeper) updateMerchantTiers(ctx sdk.Context) error {
	merchants, err := k.GetAllMerchants(ctx)
	if err != nil {
		return fmt.Errorf("failed to get merchants: %w", err)
	}
	
	for _, merchant := range merchants {
		newTier := k.calculateMerchantTier(merchant.TotalVolume)
		if newTier != merchant.Tier {
			merchant.Tier = newTier
			if err := k.SetMerchant(ctx, merchant); err != nil {
				k.logger.Error("failed to update merchant tier", 
					"merchant", merchant.Address, 
					"old_tier", merchant.Tier, 
					"new_tier", newTier,
					"error", err)
				continue
			}
			
			k.logger.Info("Updated merchant tier", 
				"merchant", merchant.Address,
				"old_tier", merchant.Tier,
				"new_tier", newTier,
				"volume", merchant.TotalVolume)
		}
	}
	
	return nil
}

// calculateMerchantTier determines merchant tier based on volume (public method)
func (k Keeper) CalculateMerchantTier(volume sdk.Coin) types.MerchantTier {
	return k.calculateMerchantTier(volume.Amount)
}

// calculateMerchantTier determines merchant tier based on volume
func (k Keeper) calculateMerchantTier(totalVolume math.Int) types.MerchantTier {
	// Define tier thresholds (in avita - smallest unit)
	platinumThreshold, _ := math.NewIntFromString("1000000000000000000000000") // 1M VITA
	goldThreshold, _ := math.NewIntFromString("100000000000000000000000")      // 100K VITA  
	silverThreshold, _ := math.NewIntFromString("10000000000000000000000")     // 10K VITA
	bronzeThreshold, _ := math.NewIntFromString("1000000000000000000000")      // 1K VITA
	
	if totalVolume.GTE(platinumThreshold) {
		return types.MerchantTierPlatinum
	} else if totalVolume.GTE(goldThreshold) {
		return types.MerchantTierGold
	} else if totalVolume.GTE(silverThreshold) {
		return types.MerchantTierSilver
	} else if totalVolume.GTE(bronzeThreshold) {
		return types.MerchantTierBronze
	}
	
	return types.MerchantTierBronze // Default tier
}

// expireOldPayments marks old pending payments as failed
func (k Keeper) expireOldPayments(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get params: %w", err)
	}
	
	payments, err := k.GetAllPayments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get payments: %w", err)
	}
	
	currentHeight := ctx.BlockHeight()
	expiredCount := 0
	
	for _, payment := range payments {
		if payment.Status == types.PaymentStatusPending {
			expirationHeight := payment.CreationHeight + int64(params.PaymentTimeoutBlocks)
			if currentHeight >= expirationHeight {
				payment.Status = types.PaymentStatusFailed
				payment.CompletionHeight = currentHeight
				
				if err := k.SetPayment(ctx, payment); err != nil {
					k.logger.Error("failed to expire payment", 
						"payment_id", payment.Id,
						"error", err)
					continue
				}
				
				expiredCount++
				k.logger.Info("Expired payment", 
					"payment_id", payment.Id,
					"creation_height", payment.CreationHeight,
					"expiration_height", expirationHeight,
					"current_height", currentHeight)
			}
		}
	}
	
	if expiredCount > 0 {
		k.logger.Info("Expired payments", "count", expiredCount, "height", currentHeight)
	}
	
	return nil
}

// processVaultUnlocks processes vault unlocks that have reached their unlock height
func (k Keeper) processVaultUnlocks(ctx sdk.Context) error {
	vaults, err := k.GetAllVaults(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vaults: %w", err)
	}
	
	currentHeight := ctx.BlockHeight()
	unlockedCount := 0
	
	for _, vault := range vaults {
		// TODO: Fix this when IsActive field is properly implemented
		// if vault.IsActive && currentHeight >= vault.UnlockHeight {
		if currentHeight >= vault.UnlockHeight {
			// Vault is now eligible for withdrawal
			// We don't automatically withdraw, but we can emit an event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeVaultUnlocked,
					sdk.NewAttribute(types.AttributeKeyVaultId, vault.Id),
					sdk.NewAttribute(types.AttributeKeyOwner, vault.Owner),
					sdk.NewAttribute(types.AttributeKeyAmount, vault.Amount.String()),
					sdk.NewAttribute(types.AttributeKeyUnlockHeight, fmt.Sprintf("%d", vault.UnlockHeight)),
				),
			)
			
			unlockedCount++
			k.logger.Info("Vault unlocked", 
				"vault_id", vault.Id,
				"owner", vault.Owner,
				"amount", vault.Amount,
				"unlock_height", vault.UnlockHeight)
		}
	}
	
	if unlockedCount > 0 {
		k.logger.Info("Vaults unlocked", "count", unlockedCount, "height", currentHeight)
	}
	
	return nil
}

// processRewardPoolStatus updates reward pool active status based on height ranges
func (k Keeper) processRewardPoolStatus(ctx sdk.Context) error {
	pools, err := k.GetAllRewardPools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reward pools: %w", err)
	}
	
	currentHeight := ctx.BlockHeight()
	statusChanges := 0
	
	for _, pool := range pools {
		shouldBeActive := currentHeight >= pool.StartHeight && 
			(pool.EndHeight == 0 || currentHeight < pool.EndHeight)
		
		if pool.IsActive != shouldBeActive {
			pool.IsActive = shouldBeActive
			if err := k.SetRewardPool(ctx, pool); err != nil {
				k.logger.Error("failed to update reward pool status", 
					"pool_id", pool.Id,
					"error", err)
				continue
			}
			
			statusChanges++
			k.logger.Info("Updated reward pool status", 
				"pool_id", pool.Id,
				"active", shouldBeActive,
				"start_height", pool.StartHeight,
				"end_height", pool.EndHeight,
				"current_height", currentHeight)
		}
	}
	
	if statusChanges > 0 {
		k.logger.Info("Reward pool status changes", "count", statusChanges, "height", currentHeight)
	}
	
	return nil
}

// calculateRewardMultiplier calculates the reward multiplier based on lock duration
// Longer lock durations receive higher multipliers to incentivize longer commitments
func (k Keeper) calculateRewardMultiplier(lockDuration uint64) math.LegacyDec {
	// Base multiplier of 1.0 (100%)
	baseMultiplier := math.LegacyNewDec(1)
	
	// Add bonus based on lock duration (in blocks)
	// Assuming ~6 second block times:
	// - 1 day (14,400 blocks): 1.1x multiplier
	// - 1 week (100,800 blocks): 1.25x multiplier  
	// - 1 month (432,000 blocks): 1.5x multiplier
	// - 1 year (5,256,000 blocks): 2.0x multiplier
	
	if lockDuration >= 5256000 { // 1 year or more
		return math.LegacyNewDecWithPrec(200, 2) // 2.0x
	} else if lockDuration >= 432000 { // 1 month or more
		return math.LegacyNewDecWithPrec(150, 2) // 1.5x
	} else if lockDuration >= 100800 { // 1 week or more
		return math.LegacyNewDecWithPrec(125, 2) // 1.25x
	} else if lockDuration >= 14400 { // 1 day or more
		return math.LegacyNewDecWithPrec(110, 2) // 1.1x
	}
	
	return baseMultiplier // 1.0x for shorter durations
}

// calculateVaultRewards calculates the reward amount for a vault based on locked amount and duration
// Rewards are calculated using the vault's reward multiplier
func (k Keeper) calculateVaultRewards(lockedAmount math.Int, lockDuration int64) math.Int {
	if lockDuration <= 0 {
		return math.ZeroInt()
	}
	
	// Convert int64 lockDuration to uint64 for calculateRewardMultiplier
	multiplier := k.calculateRewardMultiplier(uint64(lockDuration))
	
	// Calculate rewards: lockedAmount * (multiplier - 1.0)
	// e.g., for 1.1x multiplier, rewards = lockedAmount * 0.1
	oneDecimal := math.LegacyNewDec(1)
	bonusRate := multiplier.Sub(oneDecimal)
	
	// Convert lockedAmount to decimal, multiply by bonus rate, then back to Int
	lockedAmountDec := math.LegacyNewDecFromInt(lockedAmount)
	rewardsDec := lockedAmountDec.Mul(bonusRate)
	
	return rewardsDec.TruncateInt()
}