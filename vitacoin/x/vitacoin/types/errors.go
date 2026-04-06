package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module error codes
const (
	errCodeInvalidMerchant = iota + 2 // Start at 2 to avoid conflicts with sdk errors
	errCodeMerchantExists
	errCodeMerchantNotFound
	errCodeInsufficientStake
	errCodeInvalidPayment
	errCodePaymentNotFound
	errCodePaymentExpired
	errCodePaymentAlreadyCompleted
	errCodeInvalidVault
	errCodeVaultNotFound
	errCodeVaultLocked
	errCodeInsufficientVaultBalance
	errCodeInvalidRewardPool
	errCodeRewardPoolNotFound
	errCodeInsufficientPoolBalance
	errCodePoolNotActive
	errCodeInvalidTier
	errCodeUnauthorized
	errCodeInvalidParams

	// Phase 6: IBC error codes
	errCodeIBCInvalidSender
	errCodeIBCInvalidReceiver
	errCodeIBCInvalidDenom
	errCodeIBCInvalidAmount
	errCodeIBCMemoTooLong
	errCodeIBCPacketNotFound
	errCodeIBCChannelNotFound
)

// Module sentinel errors
var (
	ErrInvalidMerchant          = sdkerrors.Register(ModuleName, errCodeInvalidMerchant, "invalid merchant")
	ErrMerchantExists           = sdkerrors.Register(ModuleName, errCodeMerchantExists, "merchant already exists")
	ErrMerchantNotFound         = sdkerrors.Register(ModuleName, errCodeMerchantNotFound, "merchant not found")
	ErrInsufficientStake        = sdkerrors.Register(ModuleName, errCodeInsufficientStake, "insufficient stake amount")
	ErrInvalidPayment           = sdkerrors.Register(ModuleName, errCodeInvalidPayment, "invalid payment")
	ErrPaymentNotFound          = sdkerrors.Register(ModuleName, errCodePaymentNotFound, "payment not found")
	ErrPaymentExpired           = sdkerrors.Register(ModuleName, errCodePaymentExpired, "payment has expired")
	ErrPaymentAlreadyCompleted  = sdkerrors.Register(ModuleName, errCodePaymentAlreadyCompleted, "payment already completed")
	ErrInvalidVault             = sdkerrors.Register(ModuleName, errCodeInvalidVault, "invalid vault")
	ErrVaultNotFound            = sdkerrors.Register(ModuleName, errCodeVaultNotFound, "vault not found")
	ErrVaultLocked              = sdkerrors.Register(ModuleName, errCodeVaultLocked, "vault is still locked")
	ErrInsufficientVaultBalance = sdkerrors.Register(ModuleName, errCodeInsufficientVaultBalance, "insufficient vault balance")
	ErrInvalidRewardPool        = sdkerrors.Register(ModuleName, errCodeInvalidRewardPool, "invalid reward pool")
	ErrRewardPoolNotFound       = sdkerrors.Register(ModuleName, errCodeRewardPoolNotFound, "reward pool not found")
	ErrInsufficientPoolBalance  = sdkerrors.Register(ModuleName, errCodeInsufficientPoolBalance, "insufficient pool balance")
	ErrPoolNotActive            = sdkerrors.Register(ModuleName, errCodePoolNotActive, "reward pool is not active")
	ErrInvalidTier              = sdkerrors.Register(ModuleName, errCodeInvalidTier, "invalid merchant tier")
	ErrUnauthorized             = sdkerrors.Register(ModuleName, errCodeUnauthorized, "unauthorized")
	ErrInvalidParams            = sdkerrors.Register(ModuleName, errCodeInvalidParams, "invalid parameters")

	// Phase 6: IBC errors
	ErrInvalidSender    = sdkerrors.Register(ModuleName, errCodeIBCInvalidSender, "invalid IBC sender address")
	ErrInvalidReceiver  = sdkerrors.Register(ModuleName, errCodeIBCInvalidReceiver, "invalid IBC receiver address")
	ErrInvalidDenom     = sdkerrors.Register(ModuleName, errCodeIBCInvalidDenom, "invalid IBC token denomination")
	ErrInvalidAmount    = sdkerrors.Register(ModuleName, errCodeIBCInvalidAmount, "IBC transfer amount must be positive")
	ErrMemoTooLong      = sdkerrors.Register(ModuleName, errCodeIBCMemoTooLong, "IBC memo exceeds maximum length")
	ErrIBCPacketNotFound   = sdkerrors.Register(ModuleName, errCodeIBCPacketNotFound, "IBC pending packet not found")
	ErrIBCChannelNotFound  = sdkerrors.Register(ModuleName, errCodeIBCChannelNotFound, "IBC channel not found")
)