package keeper

import (
	"context"
	"fmt"
	
	"cosmossdk.io/math"
	
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// GetParams gets the parameters for the vitacoin module with error handling
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, fmt.Errorf("failed to get params from store: %w", err)
	}
	
	// If no params exist, return defaults
	if bz == nil {
		k.logger.Info("no params found in store, returning defaults")
		return types.DefaultParams(), nil
	}
	
	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.Params{}, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	
	return params, nil
}

// SetParams sets the parameters for the vitacoin module with full validation
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	// Validate params before setting
	if err := params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	
	if err := store.Set(types.ParamsKey, bz); err != nil {
		return fmt.Errorf("failed to set params in store: %w", err)
	}
	
	k.logger.Info("params updated",
		"min_gas_price", params.MinGasPrice.String(),
		"transaction_fee_percent", params.TransactionFeePercent.String(),
		"merchant_fee_discount", params.MerchantFeeDiscount.String(),
	)
	
	return nil
}

// GetMinGasPrice returns the minimum gas price parameter
func (k Keeper) GetMinGasPrice(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	return params.MinGasPrice, nil
}

// GetTransactionFeePercent returns the transaction fee percentage parameter
func (k Keeper) GetTransactionFeePercent(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	return params.TransactionFeePercent, nil
}

// GetMerchantFeeDiscount returns the merchant fee discount parameter
func (k Keeper) GetMerchantFeeDiscount(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	return params.MerchantFeeDiscount, nil
}

// GetMaxTransactionAmount returns the maximum transaction amount parameter
func (k Keeper) GetMaxTransactionAmount(ctx context.Context) (math.Int, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.Int{}, err
	}
	return params.MaxTransactionAmount, nil
}

// GetPaymentTimeoutBlocks returns the payment timeout in blocks parameter
func (k Keeper) GetPaymentTimeoutBlocks(ctx context.Context) (uint64, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, err
	}
	return params.PaymentTimeoutBlocks, nil
}

// GetMerchantRegistrationFee returns the merchant registration fee parameter
func (k Keeper) GetMerchantRegistrationFee(ctx context.Context) (math.Int, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.Int{}, err
	}
	return params.MerchantRegistrationFee, nil
}

// GetEnableMerchantLoyalty returns the enable merchant loyalty flag
func (k Keeper) GetEnableMerchantLoyalty(ctx context.Context) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}
	return params.EnableMerchantLoyalty, nil
}

// GetLoyaltyRewardPercent returns the loyalty reward percentage parameter
func (k Keeper) GetLoyaltyRewardPercent(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	return params.LoyaltyRewardPercent, nil
}

// GetMinMerchantStake returns the minimum merchant stake parameter
func (k Keeper) GetMinMerchantStake(ctx context.Context) (math.Int, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.Int{}, err
	}
	return params.MinMerchantStake, nil
}

// GetEnableInstantSettlement returns the enable instant settlement flag
func (k Keeper) GetEnableInstantSettlement(ctx context.Context) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}
	return params.EnableInstantSettlement, nil
}

// GetFeeBurnPercent returns the fee burn percentage parameter
func (k Keeper) GetFeeBurnPercent(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	return params.FeeBurnPercent, nil
}

// UpdateMinGasPrice updates only the minimum gas price parameter
func (k Keeper) UpdateMinGasPrice(ctx context.Context, minGasPrice math.LegacyDec) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	
	params.MinGasPrice = minGasPrice
	return k.SetParams(ctx, params)
}

// UpdateTransactionFeePercent updates only the transaction fee percentage
func (k Keeper) UpdateTransactionFeePercent(ctx context.Context, feePercent math.LegacyDec) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	
	params.TransactionFeePercent = feePercent
	return k.SetParams(ctx, params)
}

// UpdateMerchantFeeDiscount updates only the merchant fee discount
func (k Keeper) UpdateMerchantFeeDiscount(ctx context.Context, discount math.LegacyDec) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	
	params.MerchantFeeDiscount = discount
	return k.SetParams(ctx, params)
}

// ValidateParams validates all parameters with comprehensive checks
func (k Keeper) ValidateParams(params types.Params) error {
	// Validate min gas price (must be non-negative)
	if params.MinGasPrice.IsNegative() {
		return fmt.Errorf("min gas price cannot be negative: %s", params.MinGasPrice)
	}
	
	// Validate transaction fee percent (must be between 0 and 100)
	if params.TransactionFeePercent.IsNegative() || params.TransactionFeePercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("transaction fee percent must be between 0 and 100: %s", params.TransactionFeePercent)
	}
	
	// Validate merchant fee discount (must be between 0 and 100)
	if params.MerchantFeeDiscount.IsNegative() || params.MerchantFeeDiscount.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("merchant fee discount must be between 0 and 100: %s", params.MerchantFeeDiscount)
	}
	
	// Validate max transaction amount (must be non-negative)
	if params.MaxTransactionAmount.IsNegative() {
		return fmt.Errorf("max transaction amount cannot be negative: %s", params.MaxTransactionAmount)
	}
	
	// Validate payment timeout blocks (must be positive)
	if params.PaymentTimeoutBlocks == 0 {
		return fmt.Errorf("payment timeout blocks must be positive")
	}
	
	// Validate merchant registration fee (must be non-negative)
	if params.MerchantRegistrationFee.IsNegative() {
		return fmt.Errorf("merchant registration fee cannot be negative: %s", params.MerchantRegistrationFee)
	}
	
	// Validate loyalty reward percent (must be between 0 and 100)
	if params.LoyaltyRewardPercent.IsNegative() || params.LoyaltyRewardPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("loyalty reward percent must be between 0 and 100: %s", params.LoyaltyRewardPercent)
	}
	
	// Validate min merchant stake (must be non-negative)
	if params.MinMerchantStake.IsNegative() {
		return fmt.Errorf("min merchant stake cannot be negative: %s", params.MinMerchantStake)
	}
	
	// Validate fee burn percent (must be between 0 and 100)
	if params.FeeBurnPercent.IsNegative() || params.FeeBurnPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("fee burn percent must be between 0 and 100: %s", params.FeeBurnPercent)
	}
	
	return nil
}

