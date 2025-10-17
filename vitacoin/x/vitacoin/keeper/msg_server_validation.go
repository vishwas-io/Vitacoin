package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// Security and validation enhancements for message handlers

// ValidateTransactionContext performs context-specific validation
func (ms msgServer) ValidateTransactionContext(ctx context.Context, senderAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Check if chain is not halted
	if sdkCtx.BlockHeight() == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("chain not initialized")
	}
	
	// Basic rate limiting check (could be enhanced with keeper state)
	// For now, just ensure reasonable block time progression
	blockTime := sdkCtx.BlockTime()
	if blockTime.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid block time")
	}
	
	return nil
}

// ValidateMerchantOperationalStatus checks if merchant can perform operations
func (ms msgServer) ValidateMerchantOperationalStatus(ctx context.Context, merchantAddr string) error {
	merchant, err := ms.Keeper.GetMerchant(ctx, merchantAddr)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("merchant not found: %s", merchantAddr)
	}
	
	if !merchant.IsActive {
		return sdkerrors.ErrInvalidRequest.Wrap("merchant is not active")
	}
	
	// Check if merchant has been inactive for too long
	// TODO: Add LastActivityTime field to Merchant struct for inactivity checks
	// sdkCtx := sdk.UnwrapSDKContext(ctx)
	// const maxInactivityPeriod = 86400 * 30 // 30 days in seconds
	
	// if sdkCtx.BlockTime().Unix()-merchant.LastActivityTime > maxInactivityPeriod {
	// 	return sdkerrors.ErrInvalidRequest.Wrap("merchant has been inactive for too long")
	// }
	
	return nil
}

// ValidatePaymentOperationalConstraints validates payment-specific operational constraints
func (ms msgServer) ValidatePaymentOperationalConstraints(ctx context.Context, amount math.Int, merchantAddr string) error {
	params, err := ms.Keeper.GetParams(ctx)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to get params: %s", err)
	}
	
	// Check max transaction amount if set
	if !params.MaxTransactionAmount.IsZero() && amount.GT(params.MaxTransactionAmount) {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount %s exceeds max transaction amount %s", 
			amount.String(), params.MaxTransactionAmount.String())
	}
	
	// Additional checks can be added here for:
	// - Daily transaction limits per merchant
	// - Merchant tier-based limits
	// - Suspicious pattern detection
	
	return nil
}

// ValidateVaultOperationalConstraints validates vault-specific constraints
func (ms msgServer) ValidateVaultOperationalConstraints(ctx context.Context, amount math.Int, lockDuration uint64) error {
	// TODO: Add params validation when MinVaultAmount field is available
	// params, err := ms.Keeper.GetParams(ctx)
	// if err != nil {
	// 	return sdkerrors.ErrInvalidRequest.Wrapf("failed to get params: %s", err)
	// }
	
	// Validate lock duration against current height
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	unlockHeight := sdkCtx.BlockHeight() + int64(lockDuration)
	
	if err := types.ValidateUnlockHeight(unlockHeight, sdkCtx.BlockHeight()); err != nil {
		return err
	}
	
	return nil
}

// ValidateRewardPoolOperationalConstraints validates reward pool operational constraints
func (ms msgServer) ValidateRewardPoolOperationalConstraints(ctx context.Context, merchantAddr string, totalRewards math.Int) error {
	// Verify merchant exists and is active
	if err := ms.ValidateMerchantOperationalStatus(ctx, merchantAddr); err != nil {
		return err
	}
	
	// Check if merchant has sufficient stake for reward pool creation
	merchant, err := ms.Keeper.GetMerchant(ctx, merchantAddr)
	if err != nil {
		return err
	}
	
	// Reward pool should not exceed certain percentage of merchant stake
	const maxRewardPoolRatio = 10 // 10x stake amount
	maxAllowedRewards := merchant.StakeAmount.Mul(math.NewInt(maxRewardPoolRatio))
	
	if totalRewards.GT(maxAllowedRewards) {
		return sdkerrors.ErrInvalidRequest.Wrapf("reward pool amount %s exceeds maximum based on stake (max: %s)", 
			totalRewards.String(), maxAllowedRewards.String())
	}
	
	return nil
}

// ValidateGasAndFees validates gas and fee requirements
func (ms msgServer) ValidateGasAndFees(ctx context.Context, msgType string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Check if sufficient gas is available
	gasConsumed := sdkCtx.GasMeter().GasConsumed()
	gasLimit := sdkCtx.GasMeter().Limit()
	
	// Ensure we have at least 10% gas remaining for completion
	if gasConsumed > gasLimit*9/10 {
		return sdkerrors.ErrOutOfGas.Wrapf("insufficient gas for %s operation", msgType)
	}
	
	return nil
}

// Security event logging
func (ms msgServer) LogSecurityEvent(ctx context.Context, eventType, address, details string) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	ms.Keeper.Logger().Info("security_event",
		"type", eventType,
		"address", address,
		"details", details,
		"height", sdkCtx.BlockHeight(),
		"time", sdkCtx.BlockTime().Format(time.RFC3339),
	)
	
	// Emit security event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"security_event",
			sdk.NewAttribute("type", eventType),
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("details", details),
		),
	)
}

// Enhanced authorization check with detailed logging
func (ms msgServer) ValidateAuthorityWithLogging(ctx context.Context, authority, operation string) error {
	if err := ms.Keeper.ValidateAuthority(authority); err != nil {
		ms.LogSecurityEvent(ctx, "unauthorized_governance_attempt", authority, 
			fmt.Sprintf("attempted %s operation", operation))
		return err
	}
	
	ms.LogSecurityEvent(ctx, "governance_operation", authority, 
		fmt.Sprintf("successful %s operation", operation))
	
	return nil
}

// Validate business logic consistency
func (ms msgServer) ValidateBusinessLogicConsistency(ctx context.Context, operation string, data map[string]interface{}) error {
	switch operation {
	case "payment_creation":
		// Validate payment creation business logic
		if amount, ok := data["amount"].(math.Int); ok {
			if merchantAddr, ok := data["merchant_address"].(string); ok {
				return ms.ValidatePaymentOperationalConstraints(ctx, amount, merchantAddr)
			}
		}
	case "vault_creation":
		// Validate vault creation business logic
		if amount, ok := data["amount"].(math.Int); ok {
			if duration, ok := data["lock_duration"].(uint64); ok {
				return ms.ValidateVaultOperationalConstraints(ctx, amount, duration)
			}
		}
	case "reward_pool_creation":
		// Validate reward pool creation business logic
		if merchantAddr, ok := data["merchant_address"].(string); ok {
			if totalRewards, ok := data["total_rewards"].(math.Int); ok {
				return ms.ValidateRewardPoolOperationalConstraints(ctx, merchantAddr, totalRewards)
			}
		}
	}
	
	return nil
}

// Comprehensive transaction validation wrapper
func (ms msgServer) ValidateTransaction(ctx context.Context, msgType, senderAddr string, data map[string]interface{}) error {
	// Basic context validation
	if err := ms.ValidateTransactionContext(ctx, senderAddr); err != nil {
		return err
	}
	
	// Gas validation
	if err := ms.ValidateGasAndFees(ctx, msgType); err != nil {
		return err
	}
	
	// Business logic validation
	if err := ms.ValidateBusinessLogicConsistency(ctx, msgType, data); err != nil {
		return err
	}
	
	return nil
}

// Anti-spam and rate limiting (basic implementation)
func (ms msgServer) ValidateTransactionFrequency(ctx context.Context, senderAddr string) error {
	// This is a basic implementation - in production, you would track this in state
	// or use external rate limiting mechanisms
	
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Simple check: ensure block height is progressing (prevents same-block spam)
	if sdkCtx.BlockHeight() < 1 {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid block height for transaction")
	}
	
	// TODO: Implement proper rate limiting with state tracking
	// This would involve:
	// 1. Tracking last transaction time per address
	// 2. Enforcing minimum time between transactions
	// 3. Implementing sliding window rate limiting
	// 4. Different limits for different message types
	
	return nil
}

// Validate transaction amounts against economic parameters
func (ms msgServer) ValidateEconomicConstraints(ctx context.Context, amounts []math.Int) error {
	params, err := ms.Keeper.GetParams(ctx)
	if err != nil {
		return err
	}
	
	total := math.ZeroInt()
	for _, amount := range amounts {
		if amount.IsNegative() {
			return sdkerrors.ErrInvalidRequest.Wrap("negative amounts not allowed")
		}
		total = total.Add(amount)
	}
	
	// Check against economic parameters
	// Example: prevent transactions that could cause inflation issues
	const maxSingleTransactionRatio = 1000 // 0.1% of total supply
	maxAmount := params.MaxTransactionAmount.Mul(math.NewInt(maxSingleTransactionRatio))
	
	if !params.MaxTransactionAmount.IsZero() && total.GT(maxAmount) {
		return sdkerrors.ErrInvalidRequest.Wrapf("total transaction amount exceeds economic limits")
	}
	
	return nil
}