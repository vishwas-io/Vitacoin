package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return Params{
		MinGasPrice:             math.LegacyNewDecWithPrec(1, 3), // 0.001 avita
		TransactionFeePercent:   math.LegacyNewDecWithPrec(1, 1), // 0.1%
		MerchantFeeDiscount:     math.LegacyNewDecWithPrec(5, 1), // 50%
		MaxTransactionAmount:    math.NewInt(0),                  // No limit
		PaymentTimeoutBlocks:    100,                             // ~10 minutes at 6s blocks
		MerchantRegistrationFee: math.NewInt(1000000000000),      // 1000 VITA (1000 * 10^9 avita)
		EnableMerchantLoyalty:   true,
		LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(1, 2), // 1%
		MinMerchantStake:        math.NewInt(10000000000000),     // 10000 VITA
		EnableInstantSettlement: true,
		FeeBurnPercent:          math.LegacyNewDecWithPrec(25, 2), // 25%
	}
}

// Validate performs basic validation of parameters
func (p Params) Validate() error {
	if p.MinGasPrice.IsNegative() {
		return fmt.Errorf("min gas price cannot be negative: %s", p.MinGasPrice)
	}

	if p.TransactionFeePercent.IsNegative() || p.TransactionFeePercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("transaction fee percent must be between 0 and 100: %s", p.TransactionFeePercent)
	}

	if p.MerchantFeeDiscount.IsNegative() || p.MerchantFeeDiscount.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("merchant fee discount must be between 0 and 100: %s", p.MerchantFeeDiscount)
	}

	if p.MaxTransactionAmount.IsNegative() {
		return fmt.Errorf("max transaction amount cannot be negative: %s", p.MaxTransactionAmount)
	}

	if p.PaymentTimeoutBlocks == 0 {
		return fmt.Errorf("payment timeout blocks must be greater than 0")
	}

	if p.MerchantRegistrationFee.IsNegative() {
		return fmt.Errorf("merchant registration fee cannot be negative: %s", p.MerchantRegistrationFee)
	}

	if p.LoyaltyRewardPercent.IsNegative() || p.LoyaltyRewardPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("loyalty reward percent must be between 0 and 100: %s", p.LoyaltyRewardPercent)
	}

	if p.MinMerchantStake.IsNegative() {
		return fmt.Errorf("min merchant stake cannot be negative: %s", p.MinMerchantStake)
	}

	if p.FeeBurnPercent.IsNegative() || p.FeeBurnPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("fee burn percent must be between 0 and 100: %s", p.FeeBurnPercent)
	}

	return nil
}

// String returns a human-readable string representation of the parameters
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  MinGasPrice:             %s
  TransactionFeePercent:   %s%%
  MerchantFeeDiscount:     %s%%
  MaxTransactionAmount:    %s
  PaymentTimeoutBlocks:    %d
  MerchantRegistrationFee: %s
  EnableMerchantLoyalty:   %t
  LoyaltyRewardPercent:    %s%%
  MinMerchantStake:        %s
  EnableInstantSettlement: %t
  FeeBurnPercent:          %s%%`,
		p.MinGasPrice,
		p.TransactionFeePercent,
		p.MerchantFeeDiscount,
		p.MaxTransactionAmount,
		p.PaymentTimeoutBlocks,
		p.MerchantRegistrationFee,
		p.EnableMerchantLoyalty,
		p.LoyaltyRewardPercent,
		p.MinMerchantStake,
		p.EnableInstantSettlement,
		p.FeeBurnPercent,
	)
}
