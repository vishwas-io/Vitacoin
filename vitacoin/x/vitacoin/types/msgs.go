package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCreateVita represents a message to create a new Vita
type MsgCreateVita struct {
	Creator sdk.AccAddress `json:"creator" yaml:"creator"`
	Name    string         `json:"name" yaml:"name"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// Route should return the name of the module
func (msg MsgCreateVita) Route() string {
	return RouterKey
}

// Type should return the action
func (msg MsgCreateVita) Type() string {
	return "CreateVita"
}

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateVita) ValidateBasic() error {
	if msg.Creator.Empty() {
		return fmt.Errorf("creator address cannot be empty")
	}
	if len(msg.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if !msg.Amount.IsAllPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg MsgCreateVita) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}

// MsgUpdateVita represents a message to update an existing Vita
type MsgUpdateVita struct {
	Creator sdk.AccAddress `json:"creator" yaml:"creator"`
	Id      uint64         `json:"id" yaml:"id"`
	Name    string         `json:"name" yaml:"name"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// Route should return the name of the module
func (msg MsgUpdateVita) Route() string {
	return RouterKey
}

// Type should return the action
func (msg MsgUpdateVita) Type() string {
	return "UpdateVita"
}

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateVita) ValidateBasic() error {
	if msg.Creator.Empty() {
		return fmt.Errorf("creator address cannot be empty")
	}
	if msg.Id == 0 {
		return fmt.Errorf("id must be greater than zero")
	}
	if len(msg.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if !msg.Amount.IsAllPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg MsgUpdateVita) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}