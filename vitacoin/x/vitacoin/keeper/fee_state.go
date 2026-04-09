package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// GetBlockFeeAccumulator retrieves the current block's fee accumulator
func (k Keeper) GetBlockFeeAccumulator(ctx context.Context) (types.BlockFeeAccumulator, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.BlockFeeAccumulatorKey)
	if err != nil {
		return types.BlockFeeAccumulator{}, err
	}
	if bz == nil {
		return types.BlockFeeAccumulator{}, fmt.Errorf("block fee accumulator not found")
	}

	var accumulator types.BlockFeeAccumulator
	if err := json.Unmarshal(bz, &accumulator); err != nil {
		return types.BlockFeeAccumulator{}, err
	}

	return accumulator, nil
}

// SetBlockFeeAccumulator stores the current block's fee accumulator
func (k Keeper) SetBlockFeeAccumulator(ctx context.Context, accumulator types.BlockFeeAccumulator) error {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := json.Marshal(&accumulator)
	if err != nil {
		return err
	}

	return store.Set(types.BlockFeeAccumulatorKey, bz)
}

// DeleteBlockFeeAccumulator removes the current block's fee accumulator
func (k Keeper) DeleteBlockFeeAccumulator(ctx context.Context) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.BlockFeeAccumulatorKey)
}

// UpdateFeeStatistics updates cumulative fee statistics
func (k Keeper) UpdateFeeStatistics(
	ctx context.Context,
	totalCollected math.Int,
	burned math.Int,
	toValidators math.Int,
	toTreasury math.Int,
	txCount uint64,
) error {
	// Get existing statistics
	stats, err := k.GetFeeStatistics(ctx)
	if err != nil {
		// Initialize new statistics
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		stats = types.FeeStatistics{
			TotalCollectedAllTime:     math.ZeroInt(),
			TotalBurnedAllTime:        math.ZeroInt(),
			TotalToValidatorsAllTime:  math.ZeroInt(),
			TotalToTreasuryAllTime:    math.ZeroInt(),
			TotalTransactionsAllTime:  0,
			LastUpdateHeight:          sdkCtx.BlockHeight(),
			CurrentEpoch:              k.CalculateEpoch(ctx),
		}
	}

	// Update cumulative totals
	stats.TotalCollectedAllTime = stats.TotalCollectedAllTime.Add(totalCollected)
	stats.TotalBurnedAllTime = stats.TotalBurnedAllTime.Add(burned)
	stats.TotalToValidatorsAllTime = stats.TotalToValidatorsAllTime.Add(toValidators)
	stats.TotalToTreasuryAllTime = stats.TotalToTreasuryAllTime.Add(toTreasury)
	stats.TotalTransactionsAllTime += txCount
	
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	stats.LastUpdateHeight = sdkCtx.BlockHeight()
	stats.CurrentEpoch = k.CalculateEpoch(ctx)

	// Store updated statistics
	return k.SetFeeStatistics(ctx, stats)
}

// GetFeeStatistics retrieves cumulative fee statistics
func (k Keeper) GetFeeStatistics(ctx context.Context) (types.FeeStatistics, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.FeeStatisticsKey)
	if err != nil {
		return types.FeeStatistics{}, err
	}
	if bz == nil {
		return types.FeeStatistics{}, fmt.Errorf("fee statistics not found")
	}

	var stats types.FeeStatistics
	if err := json.Unmarshal(bz, &stats); err != nil {
		return types.FeeStatistics{}, err
	}

	return stats, nil
}

// SetFeeStatistics stores cumulative fee statistics
func (k Keeper) SetFeeStatistics(ctx context.Context, stats types.FeeStatistics) error {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := json.Marshal(&stats)
	if err != nil {
		return err
	}

	return store.Set(types.FeeStatisticsKey, bz)
}

// UpdateBurnStatistics updates burn-specific statistics
func (k Keeper) UpdateBurnStatistics(ctx context.Context, burnedAmount math.Int) error {
	stats, err := k.GetBurnStatistics(ctx)
	if err != nil {
		// Initialize new burn statistics
		params, paramErr := k.GetParams(ctx)
		if paramErr != nil {
			return paramErr
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		stats = types.BurnStats{
			TotalBurned:      math.ZeroInt(),
			BurnRatePerDay:   math.ZeroInt(),
			CurrentSupply:    math.NewInt(1000000000).Mul(math.NewInt(1000000000000000000)), // 1B VITA initial
			BurnCapSupply:    params.BurnCapSupply,
			RemainingToCap:   math.ZeroInt(),
			BurnCapReached:   false,
			LastBurnHeight:   sdkCtx.BlockHeight(),
		}
	}

	// Update total burned
	stats.TotalBurned = stats.TotalBurned.Add(burnedAmount)
	
	// Update current supply (decrease)
	stats.CurrentSupply = stats.CurrentSupply.Sub(burnedAmount)
	
	// Calculate remaining to cap
	if stats.CurrentSupply.GT(stats.BurnCapSupply) {
		stats.RemainingToCap = stats.CurrentSupply.Sub(stats.BurnCapSupply)
		stats.BurnCapReached = false
	} else {
		stats.RemainingToCap = math.ZeroInt()
		stats.BurnCapReached = true
	}

	// Update burn rate (approximate - needs epoch tracking for accurate daily rate)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blocksSinceLastBurn := sdkCtx.BlockHeight() - stats.LastBurnHeight
	if blocksSinceLastBurn > 0 {
		// Approximate: ~14,400 blocks per day (6 second blocks)
		blocksPerDay := int64(14400)
		if blocksSinceLastBurn < blocksPerDay {
			// Scale burned amount to daily rate
			dailyRateDec := math.LegacyNewDecFromInt(burnedAmount).Mul(
				math.LegacyNewDec(blocksPerDay).Quo(math.LegacyNewDec(blocksSinceLastBurn)),
			)
			stats.BurnRatePerDay = dailyRateDec.TruncateInt()
		}
	}
	
	stats.LastBurnHeight = sdkCtx.BlockHeight()

	return k.SetBurnStatistics(ctx, stats)
}

// GetBurnStatistics retrieves burn statistics
func (k Keeper) GetBurnStatistics(ctx context.Context) (types.BurnStats, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.BurnStatisticsKey)
	if err != nil {
		return types.BurnStats{}, err
	}
	if bz == nil {
		return types.BurnStats{}, fmt.Errorf("burn statistics not found")
	}

	var stats types.BurnStats
	if err := json.Unmarshal(bz, &stats); err != nil {
		return types.BurnStats{}, err
	}

	return stats, nil
}

// SetBurnStatistics stores burn statistics
func (k Keeper) SetBurnStatistics(ctx context.Context, stats types.BurnStats) error {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := json.Marshal(&stats)
	if err != nil {
		return err
	}

	return store.Set(types.BurnStatisticsKey, bz)
}

// CanBurnTokens checks if tokens can be burned without violating burn cap
func (k Keeper) CanBurnTokens(ctx context.Context, amount math.Int) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}

	// If burn cap is zero, no limit
	if params.BurnCapSupply.IsZero() {
		return true, nil
	}

	// Get current burn stats
	stats, err := k.GetBurnStatistics(ctx)
	if err != nil {
		// If no stats yet, allow burning
		return true, nil
	}

	// Check if already at cap
	if stats.BurnCapReached {
		return false, nil
	}

	// Check if this burn would exceed cap
	newSupply := stats.CurrentSupply.Sub(amount)
	if newSupply.LT(params.BurnCapSupply) {
		// Would go below cap - only burn up to cap
		return false, nil
	}

	return true, nil
}

// CalculateEpoch calculates the current epoch number
// 1 epoch = 1 day of blocks (~14,400 blocks at 6s/block)
func (k Keeper) CalculateEpoch(ctx context.Context) int64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blocksPerDay := int64(14400) // 24 * 60 * 60 / 6 seconds
	return sdkCtx.BlockHeight() / blocksPerDay
}

// GetSupplySnapshot retrieves supply snapshot for a specific height
func (k Keeper) GetSupplySnapshot(ctx context.Context, height int64) (types.SupplySnapshot, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	heightKey := sdk.Uint64ToBigEndian(uint64(height))
	fullKey := append(types.SupplySnapshotPrefix, heightKey...)
	
	bz, err := store.Get(fullKey)
	if err != nil {
		return types.SupplySnapshot{}, err
	}
	if bz == nil {
		return types.SupplySnapshot{}, fmt.Errorf("supply snapshot not found for height %d", height)
	}

	var snapshot types.SupplySnapshot
	if err := json.Unmarshal(bz, &snapshot); err != nil {
		return types.SupplySnapshot{}, err
	}

	return snapshot, nil
}

// SetSupplySnapshot stores a supply snapshot for a specific height
func (k Keeper) SetSupplySnapshot(ctx context.Context, snapshot types.SupplySnapshot) error {
	store := k.storeService.OpenKVStore(ctx)
	
	heightKey := sdk.Uint64ToBigEndian(uint64(snapshot.Height))
	fullKey := append(types.SupplySnapshotPrefix, heightKey...)
	
	bz, err := json.Marshal(&snapshot)
	if err != nil {
		return err
	}

	return store.Set(fullKey, bz)
}

// CreateSupplySnapshot creates a new supply snapshot for the current block
// Should be called periodically (e.g., once per epoch/day)
func (k Keeper) CreateSupplySnapshot(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Get total supply from bank keeper
	totalSupply := k.bankKeeper.GetSupply(ctx, "uvita")
	
	// Calculate circulating supply (total - locked/vested/module accounts)
	// For now, use simplified calculation
	// TODO: Implement proper circulating supply calculation considering vesting accounts
	circulatingSupply := totalSupply.Amount
	
	// Get staked amount (bonded supply)
	// TODO: Integrate with x/staking to get actual bonded amount
	bondedSupply := math.ZeroInt()
	
	// Calculate liquid supply
	liquidSupply := circulatingSupply.Sub(bondedSupply)
	
	// Get cumulative burned amount
	burnStats, err := k.GetBurnStatistics(ctx)
	burnedCumulative := math.ZeroInt()
	if err == nil {
		burnedCumulative = burnStats.TotalBurned
	}
	
	snapshot := types.SupplySnapshot{
		Height:            sdkCtx.BlockHeight(),
		Timestamp:         sdkCtx.BlockTime(),
		TotalSupply:       totalSupply.Amount,
		CirculatingSupply: circulatingSupply,
		LiquidSupply:      liquidSupply,
		BondedSupply:      bondedSupply,
		BurnedCumulative:  burnedCumulative,
	}

	return k.SetSupplySnapshot(ctx, snapshot)
}

// GetLatestSupplySnapshot retrieves the most recent supply snapshot
func (k Keeper) GetLatestSupplySnapshot(ctx context.Context) (types.SupplySnapshot, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentHeight := sdkCtx.BlockHeight()
	
	// Search backwards from current height for latest snapshot
	for height := currentHeight; height > 0 && height > currentHeight-100000; height-- {
		snapshot, err := k.GetSupplySnapshot(ctx, height)
		if err == nil {
			return snapshot, nil
		}
	}
	
	return types.SupplySnapshot{}, fmt.Errorf("no supply snapshot found")
}
