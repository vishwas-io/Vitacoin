// Package keeper — Phase 4: Validator commission & slashing logic.
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
// ValidatorRecord — stored as JSON in KV under ValidatorKeyPrefix|operatorAddr
// ---------------------------------------------------------------------------

// ValidatorRecord holds all on-chain metadata for a registered validator.
type ValidatorRecord struct {
	OperatorAddress string          `json:"operator_address"`
	Moniker         string          `json:"moniker"`
	Commission      math.LegacyDec  `json:"commission"`
	TotalDelegated  math.Int        `json:"total_delegated"`
	SelfBond        math.Int        `json:"self_bond"`
	Jailed          bool            `json:"jailed"`
	CreatedBlock    int64           `json:"created_block"`
}

// ---------------------------------------------------------------------------
// Internal KV helpers
// ---------------------------------------------------------------------------

// GetValidator retrieves a ValidatorRecord by operator address.
// Returns (record, found, error).
func (k Keeper) GetValidator(ctx context.Context, operatorAddr string) (ValidatorRecord, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetValidatorKey(operatorAddr))
	if err != nil {
		return ValidatorRecord{}, false, fmt.Errorf("get validator %s: %w", operatorAddr, err)
	}
	if bz == nil {
		return ValidatorRecord{}, false, nil
	}
	var rec ValidatorRecord
	if err := json.Unmarshal(bz, &rec); err != nil {
		return ValidatorRecord{}, false, fmt.Errorf("unmarshal validator %s: %w", operatorAddr, err)
	}
	return rec, true, nil
}

// SetValidator persists a ValidatorRecord to the KV store.
func (k Keeper) SetValidator(ctx context.Context, rec ValidatorRecord) error {
	bz, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal validator %s: %w", rec.OperatorAddress, err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetValidatorKey(rec.OperatorAddress), bz)
}

// GetAllValidators iterates the ValidatorKeyPrefix and returns every stored record.
func (k Keeper) GetAllValidators(ctx context.Context) ([]ValidatorRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.ValidatorKeyPrefix,
		storetypes.PrefixEndBytes(types.ValidatorKeyPrefix),
	)
	if err != nil {
		return nil, fmt.Errorf("open validator iterator: %w", err)
	}
	defer iter.Close()

	var records []ValidatorRecord
	for ; iter.Valid(); iter.Next() {
		var rec ValidatorRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			return nil, fmt.Errorf("unmarshal validator record: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// ---------------------------------------------------------------------------
// Public keeper methods
// ---------------------------------------------------------------------------

// RegisterValidator registers a new validator on-chain.
//
// Steps:
//  1. Validate commission is in [0.0, 1.0].
//  2. Validate selfBondAmount >= params.MinValidatorBond.
//  3. Reject if active validator count >= params.MaxValidators.
//  4. Lock self-bond tokens in the module account.
//  5. Store ValidatorRecord.
//  6. Emit EventTypeValidatorRegistered.
func (k Keeper) RegisterValidator(
	ctx context.Context,
	operatorAddr string,
	moniker string,
	commission math.LegacyDec,
	selfBondAmount math.Int,
) error {
	// 1. Validate commission [0, 1]
	if commission.IsNegative() || commission.GT(math.LegacyOneDec()) {
		return fmt.Errorf("commission must be between 0 and 1, got %s", commission)
	}

	// 2. Load staking params and validate self-bond
	params := types.DefaultStakingParams()
	if selfBondAmount.LT(params.MinValidatorBond) {
		return fmt.Errorf("self-bond %s is below minimum %s", selfBondAmount, params.MinValidatorBond)
	}

	// 3. Count existing validators
	existing, err := k.GetAllValidators(ctx)
	if err != nil {
		return fmt.Errorf("failed to count validators: %w", err)
	}
	if uint32(len(existing)) >= params.MaxValidators {
		return fmt.Errorf("validator set is full: max %d validators", params.MaxValidators)
	}

	// 4. Reject duplicate
	if _, found, err := k.GetValidator(ctx, operatorAddr); err != nil {
		return fmt.Errorf("failed to check existing validator: %w", err)
	} else if found {
		return fmt.Errorf("validator %s already registered", operatorAddr)
	}

	// 5. Lock self-bond in module account
	operatorAccAddr, err := sdk.AccAddressFromBech32(operatorAddr)
	if err != nil {
		return fmt.Errorf("invalid operator address %s: %w", operatorAddr, err)
	}
	selfBondCoin := sdk.NewCoin(types.BondDenom, selfBondAmount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		operatorAccAddr,
		types.ModuleName,
		sdk.NewCoins(selfBondCoin),
	); err != nil {
		return fmt.Errorf("failed to lock self-bond for validator %s: %w", operatorAddr, err)
	}

	// 6. Persist record
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rec := ValidatorRecord{
		OperatorAddress: operatorAddr,
		Moniker:         moniker,
		Commission:      commission,
		TotalDelegated:  selfBondAmount,
		SelfBond:        selfBondAmount,
		Jailed:          false,
		CreatedBlock:    sdkCtx.BlockHeight(),
	}
	if err := k.SetValidator(ctx, rec); err != nil {
		return fmt.Errorf("failed to store validator record: %w", err)
	}

	// 7. Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorRegistered,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, operatorAddr),
			sdk.NewAttribute(types.AttributeKeyMoniker, moniker),
			sdk.NewAttribute(types.AttributeKeyCommission, commission.String()),
			sdk.NewAttribute(types.AttributeKeySelfBond, selfBondAmount.String()),
			sdk.NewAttribute(types.AttributeKeyCreatedBlock, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
		),
	)

	k.logger.Info("registered validator",
		"operator", operatorAddr,
		"moniker", moniker,
		"commission", commission,
		"self_bond", selfBondAmount,
		"block", sdkCtx.BlockHeight(),
	)
	return nil
}

// SlashValidator slashes a validator by slashFactor of its TotalDelegated stake.
//
// slashAmount = TotalDelegated * slashFactor (truncated to Int).
// Coins are burned from the module account, and TotalDelegated is reduced accordingly.
func (k Keeper) SlashValidator(ctx context.Context, operatorAddr string, slashFactor math.LegacyDec) error {
	if slashFactor.IsNegative() || slashFactor.GT(math.LegacyOneDec()) {
		return fmt.Errorf("slash factor must be between 0 and 1, got %s", slashFactor)
	}

	rec, found, err := k.GetValidator(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to load validator %s: %w", operatorAddr, err)
	}
	if !found {
		return fmt.Errorf("validator %s not found", operatorAddr)
	}

	// Compute slash amount: TotalDelegated * slashFactor (truncate)
	slashAmountDec := math.LegacyNewDecFromInt(rec.TotalDelegated).Mul(slashFactor)
	slashAmount := slashAmountDec.TruncateInt()

	if slashAmount.IsZero() {
		return nil // nothing to slash
	}

	// Burn from module account
	slashCoin := sdk.NewCoin(types.BondDenom, slashAmount)
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(slashCoin)); err != nil {
		return fmt.Errorf("failed to burn slashed coins for validator %s: %w", operatorAddr, err)
	}

	// Update record
	rec.TotalDelegated = rec.TotalDelegated.Sub(slashAmount)
	if err := k.SetValidator(ctx, rec); err != nil {
		return fmt.Errorf("failed to update validator record after slash: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorSlashed,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, operatorAddr),
			sdk.NewAttribute(types.AttributeKeySlashFactor, slashFactor.String()),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
		),
	)

	k.logger.Info("slashed validator",
		"operator", operatorAddr,
		"slash_factor", slashFactor,
		"slash_amount", slashAmount,
		"remaining", rec.TotalDelegated,
	)
	return nil
}

// JailValidator marks the validator as jailed and persists the change.
func (k Keeper) JailValidator(ctx context.Context, operatorAddr string) error {
	rec, found, err := k.GetValidator(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to load validator %s: %w", operatorAddr, err)
	}
	if !found {
		return fmt.Errorf("validator %s not found", operatorAddr)
	}

	rec.Jailed = true
	if err := k.SetValidator(ctx, rec); err != nil {
		return fmt.Errorf("failed to jail validator %s: %w", operatorAddr, err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorJailed,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, operatorAddr),
		),
	)

	k.logger.Info("jailed validator", "operator", operatorAddr)
	return nil
}

// UnjailValidator clears the jailed flag and persists the change.
func (k Keeper) UnjailValidator(ctx context.Context, operatorAddr string) error {
	rec, found, err := k.GetValidator(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to load validator %s: %w", operatorAddr, err)
	}
	if !found {
		return fmt.Errorf("validator %s not found", operatorAddr)
	}

	rec.Jailed = false
	if err := k.SetValidator(ctx, rec); err != nil {
		return fmt.Errorf("failed to unjail validator %s: %w", operatorAddr, err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorUnjailed,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, operatorAddr),
		),
	)

	k.logger.Info("unjailed validator", "operator", operatorAddr)
	return nil
}
