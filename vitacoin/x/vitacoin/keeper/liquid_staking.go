// Package keeper implements the VitaCoin liquid staking keeper.
// Phase 4: Liquid Staking — stVITA derivative token for staked VITA.
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

// StVITADenom is the denomination for the stVITA liquid staking derivative token.
const StVITADenom = "stvita"

// ---------------------------------------------------------------------------
// stVITA Supply KV helpers
// ---------------------------------------------------------------------------

// GetStVITASupply returns the total stVITA supply from the KV store.
// Returns math.ZeroInt() if not set.
func (k Keeper) GetStVITASupply(ctx context.Context) math.Int {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetStVITASupplyKey())
	if err != nil || bz == nil {
		return math.ZeroInt()
	}
	supply := math.ZeroInt()
	if err := supply.Unmarshal(bz); err != nil {
		return math.ZeroInt()
	}
	return supply
}

// setStVITASupply stores the total stVITA supply in the KV store.
func (k Keeper) setStVITASupply(ctx context.Context, supply math.Int) error {
	bz, err := supply.Marshal()
	if err != nil {
		return fmt.Errorf("marshal stVITA supply: %w", err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetStVITASupplyKey(), bz)
}

// ---------------------------------------------------------------------------
// Exchange Rate
// ---------------------------------------------------------------------------

// GetStVITAExchangeRate returns the value of 1 stVITA expressed in VITA (avita).
// Exchange rate = totalVITADelegated / stVITASupply.
// Returns Dec(1) if either quantity is zero (initial state or empty supply).
func (k Keeper) GetStVITAExchangeRate(ctx context.Context) math.LegacyDec {
	stVITASupply := k.GetStVITASupply(ctx)
	if stVITASupply.IsZero() {
		return math.LegacyOneDec()
	}

	totalDelegated, err := k.getTotalVITADelegated(ctx)
	if err != nil || totalDelegated.IsZero() {
		return math.LegacyOneDec()
	}

	return math.LegacyNewDecFromInt(totalDelegated).Quo(math.LegacyNewDecFromInt(stVITASupply))
}

// getTotalVITADelegated sums all delegation amounts across all validators.
func (k Keeper) getTotalVITADelegated(ctx context.Context) (math.Int, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.DelegationKeyPrefix,
		storetypes.PrefixEndBytes(types.DelegationKeyPrefix),
	)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("iterate delegations: %w", err)
	}
	defer iter.Close()

	total := math.ZeroInt()
	for ; iter.Valid(); iter.Next() {
		var rec delegationRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			return math.ZeroInt(), fmt.Errorf("unmarshal delegation record: %w", err)
		}
		amt, ok := math.NewIntFromString(rec.Amount)
		if !ok {
			return math.ZeroInt(), fmt.Errorf("invalid delegation amount: %s", rec.Amount)
		}
		total = total.Add(amt)
	}
	return total, nil
}

// ---------------------------------------------------------------------------
// Mint / Burn stVITA
// ---------------------------------------------------------------------------

// MintStVITA mints stVITA tokens and sends them to toAddr.
// It also increments the tracked stVITA supply.
func (k Keeper) MintStVITA(ctx context.Context, toAddr sdk.AccAddress, vitaAmount math.Int) error {
	if toAddr.Empty() {
		return fmt.Errorf("toAddr cannot be empty")
	}
	if !vitaAmount.IsPositive() {
		return fmt.Errorf("vitaAmount must be positive, got %s", vitaAmount)
	}

	coins := sdk.NewCoins(sdk.NewCoin(StVITADenom, vitaAmount))

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("MintStVITA: mint coins: %w", err)
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, coins); err != nil {
		return fmt.Errorf("MintStVITA: send to account: %w", err)
	}

	currentSupply := k.GetStVITASupply(ctx)
	if err := k.setStVITASupply(ctx, currentSupply.Add(vitaAmount)); err != nil {
		return fmt.Errorf("MintStVITA: update supply: %w", err)
	}

	return nil
}

// BurnStVITA burns stVITA tokens from fromAddr.
// It also decrements the tracked stVITA supply.
func (k Keeper) BurnStVITA(ctx context.Context, fromAddr sdk.AccAddress, stVitaAmount math.Int) error {
	if fromAddr.Empty() {
		return fmt.Errorf("fromAddr cannot be empty")
	}
	if !stVitaAmount.IsPositive() {
		return fmt.Errorf("stVitaAmount must be positive, got %s", stVitaAmount)
	}

	coins := sdk.NewCoins(sdk.NewCoin(StVITADenom, stVitaAmount))

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fromAddr, types.ModuleName, coins); err != nil {
		return fmt.Errorf("BurnStVITA: collect from account: %w", err)
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("BurnStVITA: burn coins: %w", err)
	}

	currentSupply := k.GetStVITASupply(ctx)
	newSupply := currentSupply.Sub(stVitaAmount)
	if newSupply.IsNegative() {
		return fmt.Errorf("BurnStVITA: supply would go negative (current=%s, burn=%s)", currentSupply, stVitaAmount)
	}
	if err := k.setStVITASupply(ctx, newSupply); err != nil {
		return fmt.Errorf("BurnStVITA: update supply: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Liquid Delegate / Undelegate
// ---------------------------------------------------------------------------

// LiquidDelegate stakes VITA with a validator on behalf of the delegator
// and mints an equivalent amount of stVITA derivative tokens.
func (k Keeper) LiquidDelegate(
	ctx context.Context,
	delegatorAddr sdk.AccAddress,
	validatorAddr string,
	vitaAmount math.Int,
) error {
	if delegatorAddr.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if validatorAddr == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !vitaAmount.IsPositive() {
		return fmt.Errorf("vitaAmount must be positive, got %s", vitaAmount)
	}

	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("invalid validator address %q: %w", validatorAddr, err)
	}

	vitaCoin := sdk.NewCoin(types.BondDenom, vitaAmount)

	if err := k.DelegateToValidator(ctx, delegatorAddr, valAddr, vitaCoin); err != nil {
		return fmt.Errorf("LiquidDelegate: delegate: %w", err)
	}

	if err := k.MintStVITA(ctx, delegatorAddr, vitaAmount); err != nil {
		return fmt.Errorf("LiquidDelegate: mint stVITA: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidDelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyDelegationAmount, vitaCoin.String()),
		),
	)

	k.logger.Info("liquid delegated VITA, minted stVITA",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"vita_amount", vitaAmount,
	)
	return nil
}

// LiquidUndelegate burns stVITA tokens and initiates unbonding of the
// proportional VITA amount from the specified validator.
func (k Keeper) LiquidUndelegate(
	ctx context.Context,
	delegatorAddr sdk.AccAddress,
	validatorAddr string,
	stVitaAmount math.Int,
) error {
	if delegatorAddr.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if validatorAddr == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !stVitaAmount.IsPositive() {
		return fmt.Errorf("stVitaAmount must be positive, got %s", stVitaAmount)
	}

	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("invalid validator address %q: %w", validatorAddr, err)
	}

	exchangeRate := k.GetStVITAExchangeRate(ctx)
	// vitaToUndelegate = stVitaAmount * exchangeRate (truncated)
	vitaToUndelegateDec := math.LegacyNewDecFromInt(stVitaAmount).Mul(exchangeRate)
	vitaToUndelegate := vitaToUndelegateDec.TruncateInt()
	if !vitaToUndelegate.IsPositive() {
		return fmt.Errorf("computed VITA undelegation amount is zero for stVITA=%s rate=%s",
			stVitaAmount, exchangeRate)
	}

	vitaCoin := sdk.NewCoin(types.BondDenom, vitaToUndelegate)

	if err := k.BurnStVITA(ctx, delegatorAddr, stVitaAmount); err != nil {
		return fmt.Errorf("LiquidUndelegate: burn stVITA: %w", err)
	}

	if err := k.UndelegateFromValidator(ctx, delegatorAddr, valAddr, vitaCoin); err != nil {
		return fmt.Errorf("LiquidUndelegate: undelegate: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidUndelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute("stvita_burned", stVitaAmount.String()),
			sdk.NewAttribute(types.AttributeKeyDelegationAmount, vitaCoin.String()),
		),
	)

	k.logger.Info("liquid undelegated VITA, burned stVITA",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"stvita_burned", stVitaAmount,
		"vita_undelegated", vitaToUndelegate,
	)
	return nil
}
