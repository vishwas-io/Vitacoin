package types

import (
	"fmt"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate performs basic validation of merchant data
func (m Merchant) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return fmt.Errorf("invalid merchant address: %w", err)
	}
	
	if m.BusinessName == "" {
		return fmt.Errorf("business name cannot be empty")
	}
	
	if m.Tier < MerchantTierUnspecified || m.Tier > MerchantTierPlatinum {
		return fmt.Errorf("invalid merchant tier: %d", m.Tier)
	}
	
	if m.StakeAmount.IsNegative() {
		return fmt.Errorf("stake amount must be non-negative: %s", m.StakeAmount)
	}
	
	if m.TotalVolume.IsNegative() {
		return fmt.Errorf("total volume must be non-negative: %s", m.TotalVolume)
	}
	
	return nil
}

// Validate performs basic validation of payment data
func (p Payment) Validate() error {
	if p.Id == "" {
		return fmt.Errorf("payment ID cannot be empty")
	}
	
	if _, err := sdk.AccAddressFromBech32(p.FromAddress); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	
	if _, err := sdk.AccAddressFromBech32(p.ToAddress); err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}
	
	if p.Amount.IsNegative() || p.Amount.IsZero() {
		return fmt.Errorf("payment amount must be positive: %s", p.Amount)
	}
	
	if p.Status < PaymentStatusUnspecified || p.Status > PaymentStatusRefunded {
		return fmt.Errorf("invalid payment status: %d", p.Status)
	}
	
	if p.CreationHeight < 0 {
		return fmt.Errorf("creation height must be non-negative: %d", p.CreationHeight)
	}
	
	return nil
}

// Validate performs basic validation of vault data
func (v Vault) Validate() error {
	if v.Id == "" {
		return fmt.Errorf("vault ID cannot be empty")
	}
	
	if _, err := sdk.AccAddressFromBech32(v.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %w", err)
	}
	
	if v.Amount.IsNegative() || v.Amount.IsZero() {
		return fmt.Errorf("vault amount must be positive: %s", v.Amount)
	}
	
	if v.LockDuration == 0 {
		return fmt.Errorf("lock duration must be greater than 0")
	}
	
	if v.CreationHeight < 0 {
		return fmt.Errorf("creation height must be non-negative: %d", v.CreationHeight)
	}
	
	if v.UnlockHeight <= v.CreationHeight {
		return fmt.Errorf("unlock height must be greater than creation height")
	}
	
	if v.RewardMultiplier.IsNegative() {
		return fmt.Errorf("reward multiplier must be non-negative: %s", v.RewardMultiplier)
	}
	
	return nil
}

// Validate performs basic validation of reward pool data
func (rp RewardPool) Validate() error {
	if rp.Id == "" {
		return fmt.Errorf("pool ID cannot be empty")
	}
	
	if _, err := sdk.AccAddressFromBech32(rp.MerchantAddress); err != nil {
		return fmt.Errorf("invalid merchant address: %w", err)
	}
	
	if rp.TotalRewards.IsNegative() {
		return fmt.Errorf("total rewards must be non-negative: %s", rp.TotalRewards)
	}
	
	if rp.DistributedRewards.IsNegative() {
		return fmt.Errorf("distributed rewards must be non-negative: %s", rp.DistributedRewards)
	}
	
	if rp.DistributedRewards.GT(rp.TotalRewards) {
		return fmt.Errorf("distributed rewards cannot exceed total rewards")
	}
	
	if rp.StartHeight < 0 {
		return fmt.Errorf("start height must be non-negative: %d", rp.StartHeight)
	}
	
	if rp.EndHeight != 0 && rp.EndHeight <= rp.StartHeight {
		return fmt.Errorf("end height must be greater than start height or 0 for unlimited")
	}
	
	return nil
}

// Custom String methods for entities are handled by protobuf-generated methods
// We can add helper methods for formatting if needed in the future
