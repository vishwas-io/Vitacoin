package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// CalculateProtocolFee calculates the protocol fee for a payment amount
// Applies 0.1% fee with min/max caps as configured
// Returns (feeAmount, netAmount, error)
func (k Keeper) CalculateProtocolFee(ctx context.Context, amount math.Int) (math.Int, math.Int, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to get params: %w", err)
	}

	// If fee collection is paused, return zero fee
	if params.PausedFeeCollection {
		return math.ZeroInt(), amount, nil
	}

	// Calculate percentage-based fee
	// TransactionFeePercent is stored as a decimal fraction (e.g., 0.001 = 0.1%)
	amountDec := math.LegacyNewDecFromInt(amount)
	feeDec := amountDec.Mul(params.TransactionFeePercent)
	feeAmount := feeDec.TruncateInt()

	// Apply minimum fee cap
	if feeAmount.LT(params.MinProtocolFee) {
		feeAmount = params.MinProtocolFee
	}

	// Apply maximum fee cap
	if feeAmount.GT(params.MaxProtocolFee) {
		feeAmount = params.MaxProtocolFee
	}

	// Ensure fee doesn't exceed amount
	if feeAmount.GTE(amount) {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf(
			"calculated fee %s would exceed or equal payment amount %s (fee cannot be >= amount)",
			feeAmount, amount)
	}

	// Calculate net amount merchant receives
	netAmount := amount.Sub(feeAmount)

	return feeAmount, netAmount, nil
}

// EscrowPaymentFunds transfers payment amount from payer to module account (escrow)
// Called during MsgCreatePayment
func (k Keeper) EscrowPaymentFunds(ctx context.Context, fromAddr sdk.AccAddress, amount math.Int) error {
	if amount.IsZero() || amount.IsNegative() {
		return fmt.Errorf("invalid escrow amount: %s", amount)
	}

	// Create coins for transfer
	coins := sdk.NewCoins(sdk.NewCoin("uvita", amount))

	// Transfer from payer to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		fromAddr,
		types.ModuleName,
		coins,
	); err != nil {
		return fmt.Errorf("failed to escrow payment funds: %w", err)
	}

	k.Logger().Debug("payment funds escrowed",
		"from", fromAddr.String(),
		"amount", amount.String(),
	)

	return nil
}

// ReleasePaymentFunds releases escrowed funds on payment completion
// Calculates protocol fee, splits it according to params, and sends net to merchant
// Called during MsgCompletePayment
func (k Keeper) ReleasePaymentFunds(
	ctx context.Context,
	merchantAddr sdk.AccAddress,
	amount math.Int,
	paymentID string,
) (feeAmount math.Int, netAmount math.Int, err error) {
	// Calculate protocol fee
	feeAmount, netAmount, err = k.CalculateProtocolFee(ctx, amount)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), err
	}

	// Send net amount to merchant
	netCoins := sdk.NewCoins(sdk.NewCoin("uvita", netAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		merchantAddr,
		netCoins,
	); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to send payment to merchant: %w", err)
	}

	// If there's a protocol fee, accumulate it for end-block distribution
	if !feeAmount.IsZero() {
		if err := k.AccumulateProtocolFee(ctx, feeAmount); err != nil {
			return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to accumulate protocol fee: %w", err)
		}
	}

	// Emit event for payment settlement
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePaymentSettled,
			sdk.NewAttribute(types.AttributeKeyPaymentID, paymentID),
			sdk.NewAttribute(types.AttributeKeyMerchant, merchantAddr.String()),
			sdk.NewAttribute(types.AttributeKeyGrossAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyProtocolFee, feeAmount.String()),
			sdk.NewAttribute(types.AttributeKeyNetAmount, netAmount.String()),
		),
	)

	k.Logger().Info("payment funds released",
		"payment_id", paymentID,
		"merchant", merchantAddr.String(),
		"gross", amount.String(),
		"fee", feeAmount.String(),
		"net", netAmount.String(),
	)

	return feeAmount, netAmount, nil
}

// RefundPaymentFunds refunds escrowed funds back to payer
// Called during MsgRefundPayment
func (k Keeper) RefundPaymentFunds(ctx context.Context, toAddr sdk.AccAddress, amount math.Int, paymentID string) error {
	if amount.IsZero() || amount.IsNegative() {
		return fmt.Errorf("invalid refund amount: %s", amount)
	}

	// Transfer from module account back to payer
	coins := sdk.NewCoins(sdk.NewCoin("uvita", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		toAddr,
		coins,
	); err != nil {
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	// Emit refund event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePaymentRefunded,
			sdk.NewAttribute(types.AttributeKeyPaymentID, paymentID),
			sdk.NewAttribute(types.AttributeKeyPayer, toAddr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	k.Logger().Info("payment funds refunded",
		"payment_id", paymentID,
		"payer", toAddr.String(),
		"amount", amount.String(),
	)

	return nil
}

// AccumulateProtocolFee adds protocol fee to the current block's accumulator
// Fees are distributed in EndBlocker
func (k Keeper) AccumulateProtocolFee(ctx context.Context, amount math.Int) error {
	// Get current block accumulator
	accumulator, err := k.GetBlockFeeAccumulator(ctx)
	if err != nil {
		// If not found, create new
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		accumulator = types.BlockFeeAccumulator{
			Height:          sdkCtx.BlockHeight(),
			TotalCollected:  math.ZeroInt(),
			TransactionCount: 0,
		}
	}

	// Add to accumulator
	accumulator.TotalCollected = accumulator.TotalCollected.Add(amount)
	accumulator.TransactionCount++

	// Store updated accumulator
	if err := k.SetBlockFeeAccumulator(ctx, accumulator); err != nil {
		return fmt.Errorf("failed to set block fee accumulator: %w", err)
	}

	return nil
}

// DistributeProtocolFees distributes accumulated protocol fees according to params
// Split: 50% validators, 25% burn, 25% treasury
// Called in EndBlocker
func (k Keeper) DistributeProtocolFees(ctx context.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get params: %w", err)
	}

	// If distribution is paused, skip
	if params.PausedFeeDistribution {
		k.Logger().Info("fee distribution paused, skipping")
		return nil
	}

	// Get current block accumulator
	accumulator, err := k.GetBlockFeeAccumulator(ctx)
	if err != nil {
		// No fees collected this block
		return nil
	}

	totalFees := accumulator.TotalCollected
	if totalFees.IsZero() {
		// No fees to distribute
		return nil
	}

	// Calculate splits — percent values are already decimal fractions (e.g., 0.4 = 40%)
	burnAmount := params.FeeBurnPercent.MulInt(totalFees).TruncateInt()
	
	validatorAmount := params.FeeValidatorPercent.MulInt(totalFees).TruncateInt()
	
	treasuryAmount := params.FeeTreasuryPercent.MulInt(totalFees).TruncateInt()

	// Handle rounding - any remainder goes to validators
	distributed := burnAmount.Add(validatorAmount).Add(treasuryAmount)
	if distributed.LT(totalFees) {
		validatorAmount = validatorAmount.Add(totalFees.Sub(distributed))
	}

	// 1. Burn tokens (destroy supply)
	if !burnAmount.IsZero() {
		// Check burn cap
		canBurn, err := k.CanBurnTokens(ctx, burnAmount)
		if err != nil {
			return fmt.Errorf("failed to check burn cap: %w", err)
		}

		if canBurn {
			burnCoins := sdk.NewCoins(sdk.NewCoin("uvita", burnAmount))
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
				return fmt.Errorf("failed to burn coins: %w", err)
			}
			
			// Update burn statistics
			if err := k.UpdateBurnStatistics(ctx, burnAmount); err != nil {
				k.Logger().Error("failed to update burn statistics", "error", err)
			}
		} else {
			// If burn cap reached, redirect to treasury
			k.Logger().Info("burn cap reached, redirecting burn share to treasury",
				"amount", burnAmount.String())
			treasuryAmount = treasuryAmount.Add(burnAmount)
			burnAmount = math.ZeroInt()
		}
	}

	// 2. Send to validators via FeeCollector (x/distribution handles it)
	if !validatorAmount.IsZero() {
		validatorCoins := sdk.NewCoins(sdk.NewCoin("uvita", validatorAmount))
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			authtypes.FeeCollectorName,
			validatorCoins,
		); err != nil {
			return fmt.Errorf("failed to send fees to validators: %w", err)
		}
	}

	// 3. Send to treasury module account
	if !treasuryAmount.IsZero() {
		treasuryCoins := sdk.NewCoins(sdk.NewCoin("uvita", treasuryAmount))
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			types.TreasuryModuleName,
			treasuryCoins,
		); err != nil {
			return fmt.Errorf("failed to send fees to treasury: %w", err)
		}
	}

	// Update cumulative statistics
	if err := k.UpdateFeeStatistics(ctx, totalFees, burnAmount, validatorAmount, treasuryAmount, accumulator.TransactionCount); err != nil {
		k.Logger().Error("failed to update fee statistics", "error", err)
	}

	// Emit distribution event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFeeDistribution,
			sdk.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTotalFees, totalFees.String()),
			sdk.NewAttribute(types.AttributeKeyBurnAmount, burnAmount.String()),
			sdk.NewAttribute(types.AttributeKeyValidatorAmount, validatorAmount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryAmount, treasuryAmount.String()),
			sdk.NewAttribute(types.AttributeKeyTransactionCount, fmt.Sprintf("%d", accumulator.TransactionCount)),
		),
	)

	k.Logger().Info("protocol fees distributed",
		"total", totalFees.String(),
		"burned", burnAmount.String(),
		"validators", validatorAmount.String(),
		"treasury", treasuryAmount.String(),
		"tx_count", accumulator.TransactionCount,
	)

	// Clear block accumulator
	if err := k.DeleteBlockFeeAccumulator(ctx); err != nil {
		k.Logger().Error("failed to clear block fee accumulator", "error", err)
	}

	return nil
}
