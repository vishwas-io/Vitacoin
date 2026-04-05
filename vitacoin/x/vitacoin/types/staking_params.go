package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// ---------------------------------------------------------------------------
// Phase 4: StakingParams — separate from proto Params to avoid regen
// ---------------------------------------------------------------------------

// StakingParams defines the parameters for the VitaCoin staking system.
type StakingParams struct {
	// UnbondingTime is the duration tokens are locked after undelegation.
	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"`

	// MaxValidators is the maximum number of active validators.
	MaxValidators uint32 `json:"max_validators" yaml:"max_validators"`

	// MinValidatorBond is the minimum self-delegation a validator must maintain.
	MinValidatorBond math.Int `json:"min_validator_bond" yaml:"min_validator_bond"`

	// ValidatorCommissionRate is the default commission rate for new validators.
	ValidatorCommissionRate math.LegacyDec `json:"validator_commission_rate" yaml:"validator_commission_rate"`

	// MaxValidatorCommissionRate is the hard cap on validator commission.
	MaxValidatorCommissionRate math.LegacyDec `json:"max_validator_commission_rate" yaml:"max_validator_commission_rate"`

	// HistoricalEntries is the number of historical entries kept in state.
	HistoricalEntries uint32 `json:"historical_entries" yaml:"historical_entries"`

	// MinDelegationAmount is the minimum amount for a single delegation.
	MinDelegationAmount math.Int `json:"min_delegation_amount" yaml:"min_delegation_amount"`

	// StakingRewardRate is the annualized reward rate for stakers (e.g. 0.10 = 10% APR).
	StakingRewardRate math.LegacyDec `json:"staking_reward_rate" yaml:"staking_reward_rate"`
}

// DefaultStakingParams returns the default Phase 4 staking parameters.
func DefaultStakingParams() StakingParams {
	oneVITA := math.NewInt(1_000_000_000_000_000_000) // 1e18 avita = 1 VITA

	return StakingParams{
		UnbondingTime:              21 * 24 * time.Hour,               // 21 days (Cosmos standard)
		MaxValidators:              100,                                 // 100 active validators at launch
		MinValidatorBond:           math.NewInt(10_000).Mul(oneVITA),   // 10,000 VITA min self-stake
		ValidatorCommissionRate:    math.LegacyNewDecWithPrec(5, 2),    // 5% default commission
		MaxValidatorCommissionRate: math.LegacyNewDecWithPrec(20, 2),   // 20% hard cap
		HistoricalEntries:          10_000,
		MinDelegationAmount:        math.NewInt(1).Mul(oneVITA),        // 1 VITA minimum delegation
		StakingRewardRate:          math.LegacyNewDecWithPrec(10, 2),   // 10% APR target
	}
}

// Validate performs basic validation of staking parameters.
func (p StakingParams) Validate() error {
	if p.UnbondingTime <= 0 {
		return fmt.Errorf("unbonding time must be positive: %s", p.UnbondingTime)
	}
	if p.MaxValidators == 0 {
		return fmt.Errorf("max validators must be greater than 0")
	}
	if p.MinValidatorBond.IsNil() || p.MinValidatorBond.IsNegative() {
		return fmt.Errorf("min validator bond must be non-negative")
	}
	if p.ValidatorCommissionRate.IsNegative() || p.ValidatorCommissionRate.GT(p.MaxValidatorCommissionRate) {
		return fmt.Errorf("default commission rate must be between 0 and max (%s): %s",
			p.MaxValidatorCommissionRate, p.ValidatorCommissionRate)
	}
	if p.MaxValidatorCommissionRate.IsNegative() || p.MaxValidatorCommissionRate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("max validator commission rate must be between 0 and 1: %s", p.MaxValidatorCommissionRate)
	}
	if p.MinDelegationAmount.IsNil() || p.MinDelegationAmount.IsNegative() {
		return fmt.Errorf("min delegation amount must be non-negative")
	}
	if p.StakingRewardRate.IsNegative() {
		return fmt.Errorf("staking reward rate must be non-negative: %s", p.StakingRewardRate)
	}
	return nil
}

// String returns a human-readable representation of staking params.
func (p StakingParams) String() string {
	return fmt.Sprintf(`VitaCoin Staking Params:
  Unbonding Time:              %s
  Max Validators:              %d
  Min Validator Bond:          %s avita
  Default Commission Rate:     %s%%
  Max Commission Rate:         %s%%
  Historical Entries:          %d
  Min Delegation Amount:       %s avita
  Staking Reward Rate (APR):   %s%%`,
		p.UnbondingTime,
		p.MaxValidators,
		p.MinValidatorBond,
		p.ValidatorCommissionRate,
		p.MaxValidatorCommissionRate,
		p.HistoricalEntries,
		p.MinDelegationAmount,
		p.StakingRewardRate,
	)
}
