package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message types for routing
const (
	TypeMsgUpdateParams       = "update_params"
	TypeMsgRegisterMerchant   = "register_merchant"
	TypeMsgUpdateMerchant     = "update_merchant"
	TypeMsgCreatePayment      = "create_payment"
	TypeMsgCompletePayment    = "complete_payment"
	TypeMsgRefundPayment      = "refund_payment"
	TypeMsgCreateVault        = "create_vault"
	TypeMsgWithdrawVault      = "withdraw_vault"
	TypeMsgCreateRewardPool   = "create_reward_pool"
	TypeMsgDistributeRewards  = "distribute_rewards"
)

// ========================================
// MsgUpdateParams
// ========================================

var _ sdk.Msg = &MsgUpdateParams{}

// ValidateBasic performs stateless validation of MsgUpdateParams
func (msg *MsgUpdateParams) ValidateBasic() error {
	// Validate authority address
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	
	// Validate params
	if err := msg.Params.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid params: %s", err)
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgUpdateParams
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// ========================================
// MsgRegisterMerchant
// ========================================

var _ sdk.Msg = &MsgRegisterMerchant{}

// ValidateBasic performs stateless validation of MsgRegisterMerchant
func (msg *MsgRegisterMerchant) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate business name using enhanced validation
	if err := ValidateBusinessName(msg.BusinessName); err != nil {
		return err
	}
	
	// Validate stake amount using enhanced validation
	minStake := math.NewInt(10000000) // 10 VITA minimum (1e7 uvita) - testnet-friendly
	if err := ValidateStakeAmount(msg.StakeAmount, minStake); err != nil {
		return fmt.Errorf("invalid stake amount: %w", err)
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgRegisterMerchant
func (msg *MsgRegisterMerchant) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgUpdateMerchant
// ========================================

var _ sdk.Msg = &MsgUpdateMerchant{}

// ValidateBasic performs stateless validation of MsgUpdateMerchant
func (msg *MsgUpdateMerchant) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate business name if provided
	if msg.BusinessName != "" {
		if err := ValidateBusinessName(msg.BusinessName); err != nil {
			return err
		}
	}
	
	// Validate additional stake if provided
	if !msg.AdditionalStake.IsZero() {
		if msg.AdditionalStake.IsNegative() {
			return sdkerrors.ErrInvalidRequest.Wrap("additional stake cannot be negative")
		}
		if err := ValidatePaymentAmount(msg.AdditionalStake); err != nil {
			return fmt.Errorf("invalid additional stake: %w", err)
		}
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgUpdateMerchant
func (msg *MsgUpdateMerchant) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgCreatePayment
// ========================================

var _ sdk.Msg = &MsgCreatePayment{}

// ValidateBasic performs stateless validation of MsgCreatePayment
func (msg *MsgCreatePayment) ValidateBasic() error {
	// Validate address pair (sender and merchant)
	if err := ValidateAddressPair(msg.Sender, msg.MerchantAddress, "sender", "merchant"); err != nil {
		return err
	}
	
	// Validate payment amount using enhanced validation
	if err := ValidatePaymentAmount(msg.Amount); err != nil {
		return err
	}
	
	// Validate memo using enhanced validation
	if err := ValidateMemo(msg.Memo); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgCreatePayment
func (msg *MsgCreatePayment) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgCompletePayment
// ========================================

var _ sdk.Msg = &MsgCompletePayment{}

// ValidateBasic performs stateless validation of MsgCompletePayment
func (msg *MsgCompletePayment) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate payment ID using enhanced validation
	if err := ValidateID(msg.PaymentId, "payment"); err != nil {
		return err
	}
	
	// Security check for reentrancy patterns
	if err := ValidateNoReentrancy(msg.PaymentId); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgCompletePayment
func (msg *MsgCompletePayment) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgRefundPayment
// ========================================

var _ sdk.Msg = &MsgRefundPayment{}

// ValidateBasic performs stateless validation of MsgRefundPayment
func (msg *MsgRefundPayment) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate payment ID using enhanced validation
	if err := ValidateID(msg.PaymentId, "payment"); err != nil {
		return err
	}
	
	// Security check for reentrancy patterns
	if err := ValidateNoReentrancy(msg.PaymentId); err != nil {
		return err
	}
	
	// Validate reason using enhanced validation
	if err := ValidateReason(msg.Reason); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgRefundPayment
func (msg *MsgRefundPayment) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgCreateVault
// ========================================

var _ sdk.Msg = &MsgCreateVault{}

// ValidateBasic performs stateless validation of MsgCreateVault
func (msg *MsgCreateVault) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate vault amount using enhanced validation
	if err := ValidateVaultAmount(msg.Amount); err != nil {
		return err
	}
	
	// Validate lock duration using enhanced validation
	if err := ValidateLockDuration(msg.LockDuration); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgCreateVault
func (msg *MsgCreateVault) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgWithdrawVault
// ========================================

var _ sdk.Msg = &MsgWithdrawVault{}

// ValidateBasic performs stateless validation of MsgWithdrawVault
func (msg *MsgWithdrawVault) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate vault ID using enhanced validation
	if err := ValidateID(msg.VaultId, "vault"); err != nil {
		return err
	}
	
	// Security check for reentrancy patterns
	if err := ValidateNoReentrancy(msg.VaultId); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgWithdrawVault
func (msg *MsgWithdrawVault) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgCreateRewardPool
// ========================================

var _ sdk.Msg = &MsgCreateRewardPool{}

// ValidateBasic performs stateless validation of MsgCreateRewardPool
func (msg *MsgCreateRewardPool) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate total rewards using enhanced validation
	if err := ValidatePaymentAmount(msg.TotalRewards); err != nil {
		return fmt.Errorf("invalid total rewards: %w", err)
	}
	
	// Validate duration blocks if provided
	if msg.DurationBlocks > 0 {
		const maxDuration = MaxPoolDuration
		if msg.DurationBlocks > maxDuration {
			return sdkerrors.ErrInvalidRequest.Wrapf("duration cannot exceed %d blocks (~2 years)", maxDuration)
		}
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgCreateRewardPool
func (msg *MsgCreateRewardPool) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// ========================================
// MsgDistributeRewards
// ========================================

var _ sdk.Msg = &MsgDistributeRewards{}

// ValidateBasic performs stateless validation of MsgDistributeRewards
func (msg *MsgDistributeRewards) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	
	// Validate pool ID using enhanced validation
	if err := ValidateID(msg.PoolId, "pool"); err != nil {
		return err
	}
	
	// Security check for reentrancy patterns
	if err := ValidateNoReentrancy(msg.PoolId); err != nil {
		return err
	}
	
	// Validate reward distribution using enhanced validation
	if err := ValidateRewardDistribution(msg.Recipients, msg.Amounts); err != nil {
		return err
	}
	
	return nil
}

// GetSigners returns the expected signers for MsgDistributeRewards
func (msg *MsgDistributeRewards) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}
