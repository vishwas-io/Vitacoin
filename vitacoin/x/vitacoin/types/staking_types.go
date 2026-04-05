package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------------------------------------------------------------------------
// Phase 4 Staking Constants
// ---------------------------------------------------------------------------

const (
	TypeMsgDelegateVITA      = "delegate_vita"
	TypeMsgUndelegateVITA    = "undelegate_vita"
	TypeMsgClaimStakingRewards = "claim_staking_rewards"
	TypeMsgCreateValidator   = "create_validator"
)

// ---------------------------------------------------------------------------
// ValidatorDescription — human-readable validator metadata
// ---------------------------------------------------------------------------

// ValidatorDescription holds the description fields for a validator.
type ValidatorDescription struct {
	Moniker         string `json:"moniker" yaml:"moniker"`
	Identity        string `json:"identity" yaml:"identity"`
	Website         string `json:"website" yaml:"website"`
	SecurityContact string `json:"security_contact" yaml:"security_contact"`
	Details         string `json:"details" yaml:"details"`
}

// Validate performs basic validation of the description.
func (d ValidatorDescription) Validate() error {
	if len(d.Moniker) == 0 {
		return fmt.Errorf("moniker cannot be empty")
	}
	if len(d.Moniker) > 70 {
		return fmt.Errorf("moniker too long (max 70 chars)")
	}
	if len(d.Website) > 140 {
		return fmt.Errorf("website too long (max 140 chars)")
	}
	if len(d.Details) > 280 {
		return fmt.Errorf("details too long (max 280 chars)")
	}
	return nil
}

// ---------------------------------------------------------------------------
// CommissionRates — validator commission structure
// ---------------------------------------------------------------------------

// CommissionRates defines the initial commission rates and constraints.
type CommissionRates struct {
	Rate          math.LegacyDec `json:"rate" yaml:"rate"`
	MaxRate       math.LegacyDec `json:"max_rate" yaml:"max_rate"`
	MaxChangeRate math.LegacyDec `json:"max_change_rate" yaml:"max_change_rate"`
}

// Validate ensures commission rates are within valid bounds.
func (c CommissionRates) Validate() error {
	if c.MaxRate.IsNegative() || c.MaxRate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("max commission rate must be between 0 and 1: %s", c.MaxRate)
	}
	if c.Rate.IsNegative() || c.Rate.GT(c.MaxRate) {
		return fmt.Errorf("commission rate must be between 0 and max rate (%s): %s", c.MaxRate, c.Rate)
	}
	if c.MaxChangeRate.IsNegative() || c.MaxChangeRate.GT(c.MaxRate) {
		return fmt.Errorf("max change rate must be between 0 and max rate (%s): %s", c.MaxRate, c.MaxChangeRate)
	}
	return nil
}

// ---------------------------------------------------------------------------
// DelegationRecord — stored state for a delegation
// ---------------------------------------------------------------------------

// DelegationRecord tracks a delegator's stake with a validator.
type DelegationRecord struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           math.Int       `json:"amount" yaml:"amount"`
	StartBlock       int64          `json:"start_block" yaml:"start_block"`
	AccruedRewards   math.Int       `json:"accrued_rewards" yaml:"accrued_rewards"`
}

// ---------------------------------------------------------------------------
// UnbondingEntry — tracks tokens being unstaked
// ---------------------------------------------------------------------------

// UnbondingEntry represents a single unbonding operation.
type UnbondingEntry struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           math.Int       `json:"amount" yaml:"amount"`
	CompletionTime   time.Time      `json:"completion_time" yaml:"completion_time"`
}

// ---------------------------------------------------------------------------
// MsgDelegateVITA — delegate tokens to a validator
// ---------------------------------------------------------------------------

// MsgDelegateVITA is the message type for staking VITA with a validator.
type MsgDelegateVITA struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           sdk.Coin       `json:"amount" yaml:"amount"`
}

// NewMsgDelegateVITA creates a new MsgDelegateVITA instance.
func NewMsgDelegateVITA(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgDelegateVITA {
	return &MsgDelegateVITA{
		DelegatorAddress: delegator,
		ValidatorAddress: validator,
		Amount:           amount,
	}
}

func (msg MsgDelegateVITA) Route() string { return RouterKey }
func (msg MsgDelegateVITA) Type() string  { return TypeMsgDelegateVITA }

// ValidateBasic runs stateless validation.
func (msg MsgDelegateVITA) ValidateBasic() error {
	if msg.DelegatorAddress.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if msg.ValidatorAddress.Empty() {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return fmt.Errorf("delegation amount must be a valid positive coin: %s", msg.Amount)
	}
	if msg.Amount.Denom != BondDenom {
		return fmt.Errorf("delegation must use bond denom %s, got %s", BondDenom, msg.Amount.Denom)
	}
	return nil
}

// GetSigners returns the required signers for this message.
func (msg MsgDelegateVITA) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// ProtoMessage implements proto.Message (no-op for manual types)
func (msg *MsgDelegateVITA) ProtoMessage() {}
func (msg *MsgDelegateVITA) Reset()        {}
func (msg *MsgDelegateVITA) String() string {
	return fmt.Sprintf("MsgDelegateVITA{Delegator:%s, Validator:%s, Amount:%s}",
		msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount)
}

// ---------------------------------------------------------------------------
// MsgUndelegateVITA — begin unbonding tokens from a validator
// ---------------------------------------------------------------------------

// MsgUndelegateVITA is the message type for unbonding VITA from a validator.
type MsgUndelegateVITA struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           sdk.Coin       `json:"amount" yaml:"amount"`
}

// NewMsgUndelegateVITA creates a new MsgUndelegateVITA instance.
func NewMsgUndelegateVITA(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgUndelegateVITA {
	return &MsgUndelegateVITA{
		DelegatorAddress: delegator,
		ValidatorAddress: validator,
		Amount:           amount,
	}
}

func (msg MsgUndelegateVITA) Route() string { return RouterKey }
func (msg MsgUndelegateVITA) Type() string  { return TypeMsgUndelegateVITA }

// ValidateBasic runs stateless validation.
func (msg MsgUndelegateVITA) ValidateBasic() error {
	if msg.DelegatorAddress.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	if msg.ValidatorAddress.Empty() {
		return fmt.Errorf("validator address cannot be empty")
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return fmt.Errorf("unbonding amount must be a valid positive coin: %s", msg.Amount)
	}
	if msg.Amount.Denom != BondDenom {
		return fmt.Errorf("unbonding must use bond denom %s, got %s", BondDenom, msg.Amount.Denom)
	}
	return nil
}

// GetSigners returns the required signers for this message.
func (msg MsgUndelegateVITA) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

func (msg *MsgUndelegateVITA) ProtoMessage() {}
func (msg *MsgUndelegateVITA) Reset()        {}
func (msg *MsgUndelegateVITA) String() string {
	return fmt.Sprintf("MsgUndelegateVITA{Delegator:%s, Validator:%s, Amount:%s}",
		msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount)
}

// ---------------------------------------------------------------------------
// MsgClaimStakingRewards — claim accumulated staking rewards
// ---------------------------------------------------------------------------

// MsgClaimStakingRewards is the message type for claiming staking rewards.
type MsgClaimStakingRewards struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
}

// NewMsgClaimStakingRewards creates a new MsgClaimStakingRewards instance.
func NewMsgClaimStakingRewards(delegator sdk.AccAddress) *MsgClaimStakingRewards {
	return &MsgClaimStakingRewards{DelegatorAddress: delegator}
}

func (msg MsgClaimStakingRewards) Route() string { return RouterKey }
func (msg MsgClaimStakingRewards) Type() string  { return TypeMsgClaimStakingRewards }

// ValidateBasic runs stateless validation.
func (msg MsgClaimStakingRewards) ValidateBasic() error {
	if msg.DelegatorAddress.Empty() {
		return fmt.Errorf("delegator address cannot be empty")
	}
	return nil
}

// GetSigners returns the required signers for this message.
func (msg MsgClaimStakingRewards) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

func (msg *MsgClaimStakingRewards) ProtoMessage() {}
func (msg *MsgClaimStakingRewards) Reset()        {}
func (msg *MsgClaimStakingRewards) String() string {
	return fmt.Sprintf("MsgClaimStakingRewards{Delegator:%s}", msg.DelegatorAddress)
}

// ---------------------------------------------------------------------------
// MsgCreateValidator — register as a validator
// ---------------------------------------------------------------------------

// MsgCreateValidator is the message type for creating a new VITA validator.
type MsgCreateValidator struct {
	OperatorAddress    sdk.ValAddress       `json:"operator_address" yaml:"operator_address"`
	Description        ValidatorDescription `json:"description" yaml:"description"`
	Commission         CommissionRates      `json:"commission" yaml:"commission"`
	MinSelfDelegation  math.Int             `json:"min_self_delegation" yaml:"min_self_delegation"`
	Pubkey             string               `json:"pubkey" yaml:"pubkey"` // base64-encoded consensus pubkey
	InitialDelegation  sdk.Coin             `json:"initial_delegation" yaml:"initial_delegation"`
}

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
func NewMsgCreateValidator(
	operator sdk.ValAddress,
	description ValidatorDescription,
	commission CommissionRates,
	minSelfDelegation math.Int,
	pubkey string,
	initialDelegation sdk.Coin,
) *MsgCreateValidator {
	return &MsgCreateValidator{
		OperatorAddress:   operator,
		Description:       description,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
		Pubkey:            pubkey,
		InitialDelegation: initialDelegation,
	}
}

func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return TypeMsgCreateValidator }

// ValidateBasic runs stateless validation.
func (msg MsgCreateValidator) ValidateBasic() error {
	if msg.OperatorAddress.Empty() {
		return fmt.Errorf("operator address cannot be empty")
	}
	if err := msg.Description.Validate(); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}
	if err := msg.Commission.Validate(); err != nil {
		return fmt.Errorf("invalid commission: %w", err)
	}
	if msg.MinSelfDelegation.IsNil() || !msg.MinSelfDelegation.IsPositive() {
		return fmt.Errorf("min self delegation must be a positive integer")
	}
	if len(msg.Pubkey) == 0 {
		return fmt.Errorf("validator pubkey cannot be empty")
	}
	if !msg.InitialDelegation.IsValid() || !msg.InitialDelegation.IsPositive() {
		return fmt.Errorf("initial delegation must be a valid positive coin: %s", msg.InitialDelegation)
	}
	if msg.InitialDelegation.Denom != BondDenom {
		return fmt.Errorf("initial delegation must use bond denom %s, got %s", BondDenom, msg.InitialDelegation.Denom)
	}
	if msg.InitialDelegation.Amount.LT(msg.MinSelfDelegation) {
		return fmt.Errorf("initial delegation (%s) must be >= min self delegation (%s)",
			msg.InitialDelegation.Amount, msg.MinSelfDelegation)
	}
	return nil
}

// GetSigners returns the required signers for this message.
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// Validator operator address as AccAddress (same underlying bytes)
	return []sdk.AccAddress{sdk.AccAddress(msg.OperatorAddress)}
}

func (msg *MsgCreateValidator) ProtoMessage() {}
func (msg *MsgCreateValidator) Reset()        {}
func (msg *MsgCreateValidator) String() string {
	return fmt.Sprintf("MsgCreateValidator{Operator:%s, Moniker:%s}",
		msg.OperatorAddress, msg.Description.Moniker)
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

// MsgDelegateVITAResponse is the response type for MsgDelegateVITA.
type MsgDelegateVITAResponse struct {
	CompletionTime time.Time `json:"completion_time" yaml:"completion_time"`
}

// MsgUndelegateVITAResponse is the response type for MsgUndelegateVITA.
type MsgUndelegateVITAResponse struct {
	CompletionTime time.Time `json:"completion_time" yaml:"completion_time"`
}

// MsgClaimStakingRewardsResponse is the response type for MsgClaimStakingRewards.
type MsgClaimStakingRewardsResponse struct {
	ClaimedAmount sdk.Coins `json:"claimed_amount" yaml:"claimed_amount"`
}

// MsgCreateValidatorResponse is the response type for MsgCreateValidator.
type MsgCreateValidatorResponse struct{}
