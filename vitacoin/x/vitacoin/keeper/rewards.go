// Package keeper — Phase 4: Staking reward distribution logic.
package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ---------------------------------------------------------------------------
// pendingRewardRecord — stored as JSON in KV under PendingRewardKeyPrefix|addr
// ---------------------------------------------------------------------------

type pendingRewardRecord struct {
	Address string `json:"address"`
	Amount  string `json:"amount"` // math.Int as decimal string
	Denom   string `json:"denom"`
}

// ---------------------------------------------------------------------------
// Internal KV helpers
// ---------------------------------------------------------------------------

func (k Keeper) getPendingRewardRecord(ctx context.Context, delegatorAddr string) (pendingRewardRecord, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetPendingRewardKey(delegatorAddr))
	if err != nil {
		return pendingRewardRecord{}, false, fmt.Errorf("get pending reward %s: %w", delegatorAddr, err)
	}
	if bz == nil {
		return pendingRewardRecord{}, false, nil
	}
	var rec pendingRewardRecord
	if err := json.Unmarshal(bz, &rec); err != nil {
		return pendingRewardRecord{}, false, fmt.Errorf("unmarshal pending reward %s: %w", delegatorAddr, err)
	}
	return rec, true, nil
}

func (k Keeper) setPendingRewardRecord(ctx context.Context, rec pendingRewardRecord) error {
	bz, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal pending reward %s: %w", rec.Address, err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetPendingRewardKey(rec.Address), bz)
}

func (k Keeper) deletePendingRewardRecord(ctx context.Context, delegatorAddr string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetPendingRewardKey(delegatorAddr))
}

// getAllDelegations iterates all delegation records across all delegators/validators.
func (k Keeper) getAllDelegations(ctx context.Context) ([]delegationRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.DelegationKeyPrefix, storetypes.PrefixEndBytes(types.DelegationKeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("open delegation iterator: %w", err)
	}
	defer iter.Close()

	var records []delegationRecord
	for ; iter.Valid(); iter.Next() {
		var rec delegationRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			return nil, fmt.Errorf("unmarshal delegation record: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// ---------------------------------------------------------------------------
// Public keeper methods
// ---------------------------------------------------------------------------

// AccrueDelegatorReward adds amount/denom to a delegator's pending reward balance.
// If a record exists with a different denom, an error is returned (single-denom rewards only).
func (k Keeper) AccrueDelegatorReward(ctx context.Context, delegatorAddr string, amount math.Int, denom string) error {
	if amount.IsNegative() {
		return fmt.Errorf("reward amount must be non-negative, got %s", amount)
	}
	if denom == "" {
		return fmt.Errorf("reward denom cannot be empty")
	}

	existing, found, err := k.getPendingRewardRecord(ctx, delegatorAddr)
	if err != nil {
		return err
	}

	var newAmount math.Int
	if found {
		if existing.Denom != denom {
			return fmt.Errorf("cannot accrue reward denom %s: existing pending reward uses denom %s", denom, existing.Denom)
		}
		prev, ok := math.NewIntFromString(existing.Amount)
		if !ok {
			return fmt.Errorf("invalid stored pending reward amount: %s", existing.Amount)
		}
		newAmount = prev.Add(amount)
	} else {
		newAmount = amount
	}

	return k.setPendingRewardRecord(ctx, pendingRewardRecord{
		Address: delegatorAddr,
		Amount:  newAmount.String(),
		Denom:   denom,
	})
}

// GetPendingRewards returns the accrued pending reward for a delegator.
// Returns (ZeroInt, "", nil) when no record is found.
func (k Keeper) GetPendingRewards(ctx context.Context, delegatorAddr string) (math.Int, string, error) {
	rec, found, err := k.getPendingRewardRecord(ctx, delegatorAddr)
	if err != nil {
		return math.ZeroInt(), "", err
	}
	if !found {
		return math.ZeroInt(), "", nil
	}

	amt, ok := math.NewIntFromString(rec.Amount)
	if !ok {
		return math.ZeroInt(), "", fmt.Errorf("invalid stored pending reward amount: %s", rec.Amount)
	}
	return amt, rec.Denom, nil
}

// ClaimDelegatorRewards sends all pending rewards for a delegator to their account,
// deletes the KV record, and emits EventTypeRewardClaim.
// Returns empty sdk.Coins (no error) if there are no pending rewards.
func (k Keeper) ClaimDelegatorRewards(ctx context.Context, delegatorAddr string) (sdk.Coins, error) {
	amt, denom, err := k.GetPendingRewards(ctx, delegatorAddr)
	if err != nil {
		return nil, err
	}
	if amt.IsZero() {
		return sdk.NewCoins(), nil
	}

	accAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid delegator address %s: %w", delegatorAddr, err)
	}

	coins := sdk.NewCoins(sdk.NewCoin(denom, amt))

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, coins); err != nil {
		return nil, fmt.Errorf("failed to send rewards to %s: %w", delegatorAddr, err)
	}

	if err := k.deletePendingRewardRecord(ctx, delegatorAddr); err != nil {
		return nil, fmt.Errorf("failed to delete pending reward record for %s: %w", delegatorAddr, err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardClaim,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyRewardAmount, coins.String()),
		),
	)

	k.logger.Info("claimed staking rewards",
		"delegator", delegatorAddr,
		"amount", coins,
	)
	return coins, nil
}

// DistributeStakingRewards distributes 10% of accumulated protocol fees to validators
// and their delegators proportionally. Called from EndBlocker or a cron-equivalent hook.
//
// Distribution logic:
//  1. Sum TotalDelegated across all non-jailed validators to get the global staked pool.
//  2. Reward pool = 10% of TotalToValidatorsAllTime cumulative fees (proxy for fee pool).
//  3. Each validator receives: rewardPool * (validatorStake / totalStake).
//  4. Validator operator earns: validatorShare * commission.
//  5. Remaining share is split among delegators proportionally to their delegation weight.
//  6. Fee statistics TotalToValidatorsAllTime is reduced by the distributed amount.
func (k Keeper) DistributeStakingRewards(ctx context.Context) error {
	// --- 1. Load fee statistics to determine the reward pool ---
	stats, err := k.GetFeeStatistics(ctx)
	if err != nil {
		// No fee stats yet; nothing to distribute
		return nil
	}

	// Reward pool = 10% of the total fees routed to validators (the accumulated pool)
	rewardPoolDec := math.LegacyNewDecFromInt(stats.TotalToValidatorsAllTime).Mul(
		math.LegacyNewDecWithPrec(10, 2),
	)
	rewardPool := rewardPoolDec.TruncateInt()
	if rewardPool.IsZero() {
		return nil
	}

	// --- 2. Load all validators, skip jailed ---
	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return fmt.Errorf("DistributeStakingRewards: load validators: %w", err)
	}

	// Compute total staked across active validators
	totalStaked := math.ZeroInt()
	for _, v := range validators {
		if v.Jailed {
			continue
		}
		totalStaked = totalStaked.Add(v.TotalDelegated)
	}
	if totalStaked.IsZero() {
		return nil // no active stake, nothing to distribute
	}

	// --- 3. Load all delegation records for proportional delegator split ---
	allDelegations, err := k.getAllDelegations(ctx)
	if err != nil {
		return fmt.Errorf("DistributeStakingRewards: load delegations: %w", err)
	}

	// Build per-validator delegation map: validatorAddr → []delegationRecord
	valDelegations := make(map[string][]delegationRecord, len(validators))
	for _, d := range allDelegations {
		valDelegations[d.ValidatorAddress] = append(valDelegations[d.ValidatorAddress], d)
	}

	totalDistributed := math.ZeroInt()

	for _, val := range validators {
		if val.Jailed {
			continue
		}
		if val.TotalDelegated.IsZero() {
			continue
		}

		// Validator's proportional share of the reward pool
		valShareDec := math.LegacyNewDecFromInt(rewardPool).
			Mul(math.LegacyNewDecFromInt(val.TotalDelegated)).
			Quo(math.LegacyNewDecFromInt(totalStaked))
		valShare := valShareDec.TruncateInt()
		if valShare.IsZero() {
			continue
		}

		// Operator commission cut
		operatorRewardDec := valShareDec.Mul(val.Commission)
		operatorReward := operatorRewardDec.TruncateInt()

		// Remainder goes to delegators
		delegatorPool := valShare.Sub(operatorReward)

		// Accrue operator reward
		if operatorReward.IsPositive() {
			if err := k.AccrueDelegatorReward(ctx, val.OperatorAddress, operatorReward, types.BondDenom); err != nil {
				k.logger.Error("failed to accrue operator reward",
					"operator", val.OperatorAddress, "amount", operatorReward, "err", err)
			}
		}

		// Distribute delegator pool proportionally
		if delegatorPool.IsPositive() {
			delegations := valDelegations[val.OperatorAddress]
			for _, del := range delegations {
				delAmt, ok := math.NewIntFromString(del.Amount)
				if !ok || delAmt.IsZero() {
					continue
				}
				delRewardDec := math.LegacyNewDecFromInt(delegatorPool).
					Mul(math.LegacyNewDecFromInt(delAmt)).
					Quo(math.LegacyNewDecFromInt(val.TotalDelegated))
				delReward := delRewardDec.TruncateInt()
				if delReward.IsZero() {
					continue
				}
				if err := k.AccrueDelegatorReward(ctx, del.DelegatorAddress, delReward, types.BondDenom); err != nil {
					k.logger.Error("failed to accrue delegator reward",
						"delegator", del.DelegatorAddress, "amount", delReward, "err", err)
				}
			}
		}

		totalDistributed = totalDistributed.Add(valShare)
	}

	if totalDistributed.IsZero() {
		return nil
	}

	// --- 4. Reduce the fee pool by the distributed amount ---
	stats.TotalToValidatorsAllTime = stats.TotalToValidatorsAllTime.Sub(totalDistributed)
	if stats.TotalToValidatorsAllTime.IsNegative() {
		stats.TotalToValidatorsAllTime = math.ZeroInt()
	}
	if err := k.SetFeeStatistics(ctx, stats); err != nil {
		return fmt.Errorf("DistributeStakingRewards: update fee stats: %w", err)
	}

	// --- 5. Emit event ---
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardDistribution,
			sdk.NewAttribute(types.AttributeKeyRewardAmount, totalDistributed.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, rewardPool.String()),
		),
	)

	k.logger.Info("distributed staking rewards",
		"reward_pool", rewardPool,
		"total_distributed", totalDistributed,
		"validators", len(validators),
	)
	return nil
}
