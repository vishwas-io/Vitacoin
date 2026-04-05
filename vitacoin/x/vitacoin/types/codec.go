package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary types and interfaces for amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "vitacoin/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&MsgRegisterMerchant{}, "vitacoin/MsgRegisterMerchant", nil)
	cdc.RegisterConcrete(&MsgUpdateMerchant{}, "vitacoin/MsgUpdateMerchant", nil)
	cdc.RegisterConcrete(&MsgCreatePayment{}, "vitacoin/MsgCreatePayment", nil)
	cdc.RegisterConcrete(&MsgCompletePayment{}, "vitacoin/MsgCompletePayment", nil)
	cdc.RegisterConcrete(&MsgRefundPayment{}, "vitacoin/MsgRefundPayment", nil)
	cdc.RegisterConcrete(&MsgCreateVault{}, "vitacoin/MsgCreateVault", nil)
	cdc.RegisterConcrete(&MsgWithdrawVault{}, "vitacoin/MsgWithdrawVault", nil)
	cdc.RegisterConcrete(&MsgCreateRewardPool{}, "vitacoin/MsgCreateRewardPool", nil)
	cdc.RegisterConcrete(&MsgDistributeRewards{}, "vitacoin/MsgDistributeRewards", nil)

	// Phase 4: Staking messages
	cdc.RegisterConcrete(&MsgDelegateVITA{}, "vitacoin/MsgDelegateVITA", nil)
	cdc.RegisterConcrete(&MsgUndelegateVITA{}, "vitacoin/MsgUndelegateVITA", nil)
	cdc.RegisterConcrete(&MsgClaimStakingRewards{}, "vitacoin/MsgClaimStakingRewards", nil)
	cdc.RegisterConcrete(&MsgCreateValidator{}, "vitacoin/MsgCreateValidator", nil)
}

// RegisterInterfaces registers the module's interfaces and implementations with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgRegisterMerchant{},
		&MsgUpdateMerchant{},
		&MsgCreatePayment{},
		&MsgCompletePayment{},
		&MsgRefundPayment{},
		&MsgCreateVault{},
		&MsgWithdrawVault{},
		&MsgCreateRewardPool{},
		&MsgDistributeRewards{},
	)

	// Phase 4: Staking message interface registrations
	// Note: These use manual Go types (no protoc); registered for amino routing only.
	// Full interface registry integration requires proto regen in a future step.

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}