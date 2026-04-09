package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultParams returns default module parameters
func DefaultParams() Params {
	oneVITA := math.NewInt(1000000) // 1e6 uvita = 1 VITA
	
	return Params{
		MinGasPrice:             math.LegacyNewDecWithPrec(25, 3), // 0.025 uvita
		TransactionFeePercent:   math.LegacyNewDecWithPrec(1, 3), // 0.1% (was 0.1 = 10%)
		MerchantFeeDiscount:     math.LegacyZeroDec(),            // 0% discount (disabled for mainnet)
		MaxTransactionAmount:    math.NewInt(0),                  // No limit
		PaymentTimeoutBlocks:    100,                             // ~10 minutes at 6s blocks
		MerchantRegistrationFee: math.NewInt(1000).Mul(oneVITA), // 1000 VITA
		EnableMerchantLoyalty:   true,
		LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(1, 2), // 1%
		MinMerchantStake:        math.NewInt(1000).Mul(oneVITA),  // 1000 VITA - Bronze tier
		EnableInstantSettlement: true,
		
		// Phase 3: Fee Distribution Parameters
		FeeBurnPercent:         math.LegacyNewDecWithPrec(40, 2),  // 40% burn
		FeeValidatorPercent:    math.LegacyNewDecWithPrec(40, 2),  // 40% to validators
		FeeTreasuryPercent:     math.LegacyNewDecWithPrec(20, 2),  // 20% to treasury
		MinProtocolFee:         math.NewInt(1000),                    // 0.001 VITA (1000 uvita)
		MaxProtocolFee:         math.NewInt(100000000),               // 100 VITA (100 * 1e6 uvita)
		BurnCapSupply:          math.NewInt(100000000000000),         // 100M VITA (100M * 1e6 uvita)
		PausedFeeCollection:    false,
		PausedFeeDistribution:  false,
		
		// Circuit Breaker: Emergency pause flags (governance-controlled)
		PausedPayments:         false,
		PausedStaking:          false,
		PausedIBC:              false,
	}
}

// Validate performs basic validation of parameters
func (p Params) Validate() error {
	if p.MinGasPrice.IsNegative() {
		return fmt.Errorf("min gas price must be non-negative: %s", p.MinGasPrice)
	}

	if p.TransactionFeePercent.IsNegative() || p.TransactionFeePercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("transaction fee percent must be between 0 and 100: %s", p.TransactionFeePercent)
	}

	if p.MerchantFeeDiscount.IsNegative() || p.MerchantFeeDiscount.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("merchant fee discount must be between 0 and 100: %s", p.MerchantFeeDiscount)
	}

	if p.MaxTransactionAmount.IsNegative() {
		return fmt.Errorf("max transaction amount must be non-negative: %s", p.MaxTransactionAmount)
	}

	if p.PaymentTimeoutBlocks == 0 {
		return fmt.Errorf("payment timeout blocks must be greater than 0")
	}

	if p.MerchantRegistrationFee.IsNegative() {
		return fmt.Errorf("merchant registration fee must be non-negative: %s", p.MerchantRegistrationFee)
	}

	if p.LoyaltyRewardPercent.IsNegative() || p.LoyaltyRewardPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("loyalty reward percent must be between 0 and 100: %s", p.LoyaltyRewardPercent)
	}

	if p.MinMerchantStake.IsNegative() {
		return fmt.Errorf("min merchant stake must be non-negative: %s", p.MinMerchantStake)
	}

	if p.FeeBurnPercent.IsNegative() || p.FeeBurnPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("fee burn percent must be between 0 and 100: %s", p.FeeBurnPercent)
	}

	// Phase 3: Validate new fee distribution parameters
	if p.FeeValidatorPercent.IsNegative() || p.FeeValidatorPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("fee validator percent must be between 0 and 100: %s", p.FeeValidatorPercent)
	}

	if p.FeeTreasuryPercent.IsNegative() || p.FeeTreasuryPercent.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("fee treasury percent must be between 0 and 100: %s", p.FeeTreasuryPercent)
	}

	// Validate total fee split equals 100% (values are fractions: 0.25 + 0.50 + 0.25 = 1.0)
	totalFeePercent := p.FeeBurnPercent.Add(p.FeeValidatorPercent).Add(p.FeeTreasuryPercent)
	onehundredPercent := math.LegacyNewDec(1)
	if !totalFeePercent.Equal(onehundredPercent) {
		return fmt.Errorf("fee split must total 100%%, got %s%% (burn: %s%%, validator: %s%%, treasury: %s%%)",
			totalFeePercent, p.FeeBurnPercent, p.FeeValidatorPercent, p.FeeTreasuryPercent)
	}

	if p.MinProtocolFee.IsNegative() {
		return fmt.Errorf("min protocol fee must be non-negative: %s", p.MinProtocolFee)
	}

	if p.MaxProtocolFee.IsNegative() {
		return fmt.Errorf("max protocol fee must be non-negative: %s", p.MaxProtocolFee)
	}

	if p.MaxProtocolFee.LT(p.MinProtocolFee) {
		return fmt.Errorf("max protocol fee %s must be >= min protocol fee %s", 
			p.MaxProtocolFee, p.MinProtocolFee)
	}

	if p.BurnCapSupply.IsNegative() {
		return fmt.Errorf("burn cap supply must be non-negative: %s", p.BurnCapSupply)
	}

	return nil
}

// String returns a human-readable string representation of the parameters
func (p Params) String() string {
	return fmt.Sprintf(`Vitacoin Params:
  Min Gas Price:             %s
  Transaction Fee Percent:   %s%%
  Merchant Fee Discount:     %s%%
  Max Transaction Amount:    %s
  Payment Timeout Blocks:    %d
  Merchant Registration Fee: %s
  Enable Merchant Loyalty:   %t
  Loyalty Reward Percent:    %s%%
  Min Merchant Stake:        %s
  Enable Instant Settlement: %t
  Fee Distribution:
    Burn Percent:            %s%%
    Validator Percent:       %s%%
    Treasury Percent:        %s%%
  Protocol Fee Limits:
    Min Protocol Fee:        %s
    Max Protocol Fee:        %s
  Burn Cap Supply:           %s
  Emergency Flags:
    Fee Collection Paused:   %t
    Fee Distribution Paused: %t
    Payments Paused:         %t
    Staking Paused:          %t
    IBC Paused:              %t`,
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
		p.FeeValidatorPercent,
		p.FeeTreasuryPercent,
		p.MinProtocolFee,
		p.MaxProtocolFee,
		p.BurnCapSupply,
		p.PausedFeeCollection,
		p.PausedFeeDistribution,
		p.PausedPayments,
		p.PausedStaking,
		p.PausedIBC,
	)
}
