// Package keeper implements the VitaCoin staking keeper.
// Phase 4: Staking System — delegate, undelegate, unbonding release.
package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// defaultUnbondingBlocks is used when params cannot be loaded.
// Corresponds to 21 days at ~6s block time.
const defaultUnbondingBlocks = int64(302400)

// ---------------------------------------------------------------------------
// Internal storage helpers
// ---------------------------------------------------------------------------

// delegationRecord is the JSON-encoded form stored in KV for a delegation.
type delegationRecord struct {
	DelegatorAddress string `json:"delegator"`
	ValidatorAddress string `json:"validator"`
	Amount           string `json:"amount"`  // math.Int as string
	Denom            string `json:"denom"`
	StartBlock       int64  `json:"start_block"`
}

// unbondingRecord is the JSON-encoded form stored in KV for an unbonding entry.
type unbondingRecord struct {
	DelegatorAddress string    `json:"delegator"`
	ValidatorAddress string    `json:"validator"`
	Amount           string    `json:"amount"` // math.Int as string
	Denom            string    `json:"denom"`
	MaturityBlock    int64     `json:"maturity_block"`
	CreatedAt        time.Time `json:"created_at"`
}

// setDelegation stores or updates a delegation record.
func (k Keeper) setDelegation(ctx context.Context, rec delegationRecord) error {
	bz, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal delegation: %w", err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetDelegationKey(rec.DelegatorAddress, rec.ValidatorAddress), bz)
}

// getDelegation retrieves a delegation record. Returns (record, found, error).
func (k Keeper) getDelegation(ctx context.Context, delegatorAddr, validatorAddr string) (delegationRecord, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetDelegationKey(delegatorAddr, validatorAddr))
	if err != nil {
		return delegationRecord{}, false, err
	}
	if bz == nil {
		return delegationRecord{}, false, nil
	}
	var rec delegationRecord
	if err := json.Unmarshal(bz, &rec); err != nil {
		return delegationRecord{}, false, fmt.Errorf("unmarshal delegation: %w", err)
	}
	return rec, true, nil
}

// deleteDelegation removes a delegation record.
func (k Keeper) deleteDelegation(ctx context.Context, delegatorAddr, validatorAddr string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetDelegationKey(delegatorAddr, validatorAddr))
}

// setUnbonding stores an unbonding record.
func (k Keeper) setUnbonding(ctx context.Context, rec unbondingRecord) error {
	bz, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal unbonding: %w", err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetUnbondingKey(rec.DelegatorAddress, rec.ValidatorAddress, rec.MaturityBlock), bz)
}

// getAllUnbondings iterates all unbonding records. Callers must not modify the store during iteration.
func (k Keeper) getAllUnbondings(ctx context.Context) ([]unbondingRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.UnbondingKeyPrefix, storetypes.PrefixEndBytes(types.UnbondingKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var records []unbondingRecord
	for ; iter.Valid(); iter.Next() {
		var rec unbondingRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			return nil, fmt.Errorf("unmarshal unbonding: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// deleteUnbonding removes an unbonding record.
func (k Keeper) deleteUnbonding(ctx context.Context, delegatorAddr, validatorAddr string, maturityBlock int64) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetUnbondingKey(delegatorAddr, validatorAddr, maturityBlock))
}

// ---------------------------------------------------------------------------
// Public keeper methods
// ---------------------------------------------------------------------------

// DelegateToValidator stakes VITA tokens with the given validator on behalf of the delegator.
//
// Steps:
//  1. Validate coin denom and amount.
//  2. Lock VITA in the module account via bankKeeper.SendCoinsFromAccountToModule.
//  3. Upsert a DelegationRecord in the KV store.
//  4. Emit EventTypeDelegation.
func (k Keeper) DelegateToValidator(
	ctx context.Context,
	delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress,
	amount sdk.Coin,
) error {
	// 1. Basic validation
	if delegatorAddr.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if validatorAddr.Empty() {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !amount.IsValid() || !amount.IsPositive() {
		return fmt.Errorf("delegation amount must be a valid positive coin: %s", amount)
	}
	if amount.Denom != types.BondDenom {
		return fmt.Errorf("delegation must use bond denom %s, got %s", types.BondDenom, amount.Denom)
	}

	// 2. Lock coins in module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		delegatorAddr,
		types.ModuleName,
		sdk.NewCoins(amount),
	); err != nil {
		return fmt.Errorf("failed to lock delegation in module account: %w", err)
	}

	// 3. Upsert delegation record (accumulate if one already exists)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	existing, found, err := k.getDelegation(ctx, delegatorAddr.String(), validatorAddr.String())
	if err != nil {
		return fmt.Errorf("failed to read existing delegation: %w", err)
	}

	var startBlock int64
	var totalAmountStr string
	if found {
		// Accumulate
		prev, ok := math.NewIntFromString(existing.Amount)
		if !ok {
			return fmt.Errorf("invalid stored delegation amount: %s", existing.Amount)
		}
		totalAmountStr = prev.Add(amount.Amount).String()
		startBlock = existing.StartBlock
	} else {
		totalAmountStr = amount.Amount.String()
		startBlock = sdkCtx.BlockHeight()
	}

	rec := delegationRecord{
		DelegatorAddress: delegatorAddr.String(),
		ValidatorAddress: validatorAddr.String(),
		Amount:           totalAmountStr,
		Denom:            amount.Denom,
		StartBlock:       startBlock,
	}
	if err := k.setDelegation(ctx, rec); err != nil {
		return fmt.Errorf("failed to store delegation record: %w", err)
	}

	// 4. Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelegation,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDelegationAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyStartBlock, fmt.Sprintf("%d", startBlock)),
		),
	)

	k.logger.Info("delegated VITA",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"amount", amount,
		"block", sdkCtx.BlockHeight(),
	)
	return nil
}

// UndelegateFromValidator initiates an unbonding of VITA tokens from the given validator.
//
// Steps:
//  1. Validate inputs and check delegation exists with sufficient balance.
//  2. Reduce (or delete) the delegation record.
//  3. Create an UnbondingRecord with maturity = currentBlock + UnbondingBlocks.
//  4. Emit EventTypeUnbonding.
//
// Tokens remain locked in the module account until ProcessMatureUnbondings releases them.
func (k Keeper) UndelegateFromValidator(
	ctx context.Context,
	delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress,
	amount sdk.Coin,
) error {
	// 1. Validate inputs
	if delegatorAddr.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if validatorAddr.Empty() {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !amount.IsValid() || !amount.IsPositive() {
		return fmt.Errorf("undelegation amount must be a valid positive coin: %s", amount)
	}
	if amount.Denom != types.BondDenom {
		return fmt.Errorf("undelegation must use bond denom %s, got %s", types.BondDenom, amount.Denom)
	}

	// 2. Load delegation
	existing, found, err := k.getDelegation(ctx, delegatorAddr.String(), validatorAddr.String())
	if err != nil {
		return fmt.Errorf("failed to read delegation: %w", err)
	}
	if !found {
		return fmt.Errorf("no delegation found for delegator %s and validator %s",
			delegatorAddr, validatorAddr)
	}

	delegatedAmt, ok := math.NewIntFromString(existing.Amount)
	if !ok {
		return fmt.Errorf("invalid stored delegation amount: %s", existing.Amount)
	}
	if amount.Amount.GT(delegatedAmt) {
		return fmt.Errorf("cannot undelegate %s, only %s delegated", amount.Amount, delegatedAmt)
	}

	// 3. Update (or remove) delegation record
	remaining := delegatedAmt.Sub(amount.Amount)
	if remaining.IsZero() {
		if err := k.deleteDelegation(ctx, delegatorAddr.String(), validatorAddr.String()); err != nil {
			return fmt.Errorf("failed to delete delegation record: %w", err)
		}
	} else {
		existing.Amount = remaining.String()
		if err := k.setDelegation(ctx, existing); err != nil {
			return fmt.Errorf("failed to update delegation record: %w", err)
		}
	}

	// 4. Create unbonding record
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Use default unbonding blocks (21 days at ~6s/block = 302400 blocks).
	// StakingParams keeper integration will be added when the full staking param
	// store is wired in Phase 4 completion.
	unbondingBlocks := defaultUnbondingBlocks
	maturityBlock := sdkCtx.BlockHeight() + unbondingBlocks

	ubRec := unbondingRecord{
		DelegatorAddress: delegatorAddr.String(),
		ValidatorAddress: validatorAddr.String(),
		Amount:           amount.Amount.String(),
		Denom:            amount.Denom,
		MaturityBlock:    maturityBlock,
		CreatedAt:        sdkCtx.BlockTime(),
	}
	if err := k.setUnbonding(ctx, ubRec); err != nil {
		return fmt.Errorf("failed to store unbonding record: %w", err)
	}

	// 5. Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonding,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDelegationAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyMaturityBlock, fmt.Sprintf("%d", maturityBlock)),
		),
	)

	k.logger.Info("initiated undelegation",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"amount", amount,
		"maturity_block", maturityBlock,
	)
	return nil
}

// ProcessMatureUnbondings is called from EndBlocker. It iterates all unbonding
// records and releases tokens for entries whose MaturityBlock has been reached.
func (k Keeper) ProcessMatureUnbondings(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentBlock := sdkCtx.BlockHeight()

	records, err := k.getAllUnbondings(ctx)
	if err != nil {
		return fmt.Errorf("failed to iterate unbondings: %w", err)
	}

	for _, rec := range records {
		if currentBlock < rec.MaturityBlock {
			continue
		}

		// Parse amount
		amt, ok := math.NewIntFromString(rec.Amount)
		if !ok {
			k.logger.Error("invalid unbonding amount, skipping", "amount", rec.Amount)
			continue
		}

		delegatorAddr, err := sdk.AccAddressFromBech32(rec.DelegatorAddress)
		if err != nil {
			k.logger.Error("invalid delegator address in unbonding", "addr", rec.DelegatorAddress, "err", err)
			continue
		}

		coin := sdk.NewCoin(rec.Denom, amt)

		// Release from module account
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			delegatorAddr,
			sdk.NewCoins(coin),
		); err != nil {
			k.logger.Error("failed to release unbonded tokens",
				"delegator", rec.DelegatorAddress,
				"amount", coin,
				"err", err,
			)
			continue
		}

		// Remove unbonding record
		if err := k.deleteUnbonding(ctx, rec.DelegatorAddress, rec.ValidatorAddress, rec.MaturityBlock); err != nil {
			k.logger.Error("failed to delete unbonding record",
				"delegator", rec.DelegatorAddress,
				"maturity_block", rec.MaturityBlock,
				"err", err,
			)
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUnbondingReleased,
				sdk.NewAttribute(types.AttributeKeyDelegator, rec.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeyValidator, rec.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegationAmount, coin.String()),
				sdk.NewAttribute(types.AttributeKeyMaturityBlock, fmt.Sprintf("%d", rec.MaturityBlock)),
			),
		)

		k.logger.Info("released mature unbonding",
			"delegator", rec.DelegatorAddress,
			"amount", coin,
			"block", currentBlock,
		)
	}

	return nil
}

// ClaimStakingRewards collects all accrued staking rewards for the delegator across all validators.
// Phase 4 stub — rewards system will be completed in a follow-up task.
func (k Keeper) ClaimStakingRewards(
	ctx context.Context,
	delegatorAddr sdk.AccAddress,
) (sdk.Coins, error) {
	return sdk.NewCoins(), nil
}

// GetValidatorAPR returns the annualised percentage rate (APR) for the given validator.
// Phase 4 stub — APR computation will be completed in a follow-up task.
func (k Keeper) GetValidatorAPR(
	ctx context.Context,
	validatorAddr sdk.ValAddress,
) (string, error) {
	return "0.0000", nil
}
