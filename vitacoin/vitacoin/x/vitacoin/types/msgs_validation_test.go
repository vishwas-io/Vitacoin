package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

func TestMsgUpdateParamsValidateBasic(t *testing.T) {
	validAuthority := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"

	tests := []struct {
		name      string
		msg       *types.MsgUpdateParams
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: &types.MsgUpdateParams{
				Authority: validAuthority,
				Params:    types.DefaultParams(),
			},
			expectErr: false,
		},
		{
			name: "invalid authority address",
			msg: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			expectErr: true,
			errMsg:    "invalid authority address",
		},
		{
			name: "invalid params - negative fee percent",
			msg: &types.MsgUpdateParams{
				Authority: validAuthority,
				Params: types.Params{
					MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
					TransactionFeePercent: math.LegacyNewDec(-1),
				},
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRegisterMerchantValidateBasic(t *testing.T) {
	validAddress := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"

	tests := []struct {
		name      string
		msg       *types.MsgRegisterMerchant
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: &types.MsgRegisterMerchant{
				Sender:       validAddress,
				BusinessName: "Test Business",
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: &types.MsgRegisterMerchant{
				Sender:       "invalid",
				BusinessName: "Test Business",
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "invalid sender address",
		},
		{
			name: "empty business name",
			msg: &types.MsgRegisterMerchant{
				Sender:       validAddress,
				BusinessName: "",
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "business name cannot be empty",
		},
		{
			name: "business name too long",
			msg: &types.MsgRegisterMerchant{
				Sender:       validAddress,
				BusinessName: string(make([]byte, 101)), // 101 characters
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "business name cannot exceed 100 characters",
		},
		{
			name: "zero stake amount",
			msg: &types.MsgRegisterMerchant{
				Sender:       validAddress,
				BusinessName: "Test Business",
				StakeAmount:  math.ZeroInt(),
			},
			expectErr: true,
			errMsg:    "stake amount must be positive",
		},
		{
			name: "negative stake amount",
			msg: &types.MsgRegisterMerchant{
				Sender:       validAddress,
				BusinessName: "Test Business",
				StakeAmount:  math.NewInt(-1000),
			},
			expectErr: true,
			errMsg:    "stake amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreatePaymentValidateBasic(t *testing.T) {
	validSender := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"
	validMerchant := "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5"

	tests := []struct {
		name      string
		msg       *types.MsgCreatePayment
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: &types.MsgCreatePayment{
				Sender:          validSender,
				MerchantAddress: validMerchant,
				Amount:          math.NewInt(1000),
				Memo:            "Test payment",
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: &types.MsgCreatePayment{
				Sender:          "invalid",
				MerchantAddress: validMerchant,
				Amount:          math.NewInt(1000),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "invalid sender address",
		},
		{
			name: "invalid merchant address",
			msg: &types.MsgCreatePayment{
				Sender:          validSender,
				MerchantAddress: "invalid",
				Amount:          math.NewInt(1000),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "invalid merchant address",
		},
		{
			name: "sender and merchant same",
			msg: &types.MsgCreatePayment{
				Sender:          validSender,
				MerchantAddress: validSender,
				Amount:          math.NewInt(1000),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "sender and merchant cannot be the same",
		},
		{
			name: "zero amount",
			msg: &types.MsgCreatePayment{
				Sender:          validSender,
				MerchantAddress: validMerchant,
				Amount:          math.ZeroInt(),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "amount must be positive",
		},
		{
			name: "memo too long",
			msg: &types.MsgCreatePayment{
				Sender:          validSender,
				MerchantAddress: validMerchant,
				Amount:          math.NewInt(1000),
				Memo:            string(make([]byte, 257)), // 257 characters
			},
			expectErr: true,
			errMsg:    "memo cannot exceed 256 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateVaultValidateBasic(t *testing.T) {
	validAddress := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"

	tests := []struct {
		name      string
		msg       *types.MsgCreateVault
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: &types.MsgCreateVault{
				Sender:       validAddress,
				Amount:       math.NewInt(5000),
				LockDuration: 1000,
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: &types.MsgCreateVault{
				Sender:       "invalid",
				Amount:       math.NewInt(5000),
				LockDuration: 1000,
			},
			expectErr: true,
			errMsg:    "invalid sender address",
		},
		{
			name: "zero amount",
			msg: &types.MsgCreateVault{
				Sender:       validAddress,
				Amount:       math.ZeroInt(),
				LockDuration: 1000,
			},
			expectErr: true,
			errMsg:    "amount must be positive",
		},
		{
			name: "zero lock duration",
			msg: &types.MsgCreateVault{
				Sender:       validAddress,
				Amount:       math.NewInt(5000),
				LockDuration: 0,
			},
			expectErr: true,
			errMsg:    "lock duration must be greater than 0",
		},
		{
			name: "lock duration too long",
			msg: &types.MsgCreateVault{
				Sender:       validAddress,
				Amount:       math.NewInt(5000),
				LockDuration: 10_000_000, // More than max
			},
			expectErr: true,
			errMsg:    "lock duration cannot exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgDistributeRewardsValidateBasic(t *testing.T) {
	validSender := "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5"
	validRecipient1 := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"
	validRecipient2 := "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5"

	tests := []struct {
		name      string
		msg       *types.MsgDistributeRewards
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{validRecipient1, validRecipient2},
				Amounts:    []math.Int{math.NewInt(100), math.NewInt(200)},
			},
			expectErr: false,
		},
		{
			name: "invalid sender address",
			msg: &types.MsgDistributeRewards{
				Sender:     "invalid",
				PoolId:     "pool-1",
				Recipients: []string{validRecipient1},
				Amounts:    []math.Int{math.NewInt(100)},
			},
			expectErr: true,
			errMsg:    "invalid sender address",
		},
		{
			name: "empty pool ID",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "",
				Recipients: []string{validRecipient1},
				Amounts:    []math.Int{math.NewInt(100)},
			},
			expectErr: true,
			errMsg:    "pool ID cannot be empty",
		},
		{
			name: "empty recipients list",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{},
				Amounts:    []math.Int{},
			},
			expectErr: true,
			errMsg:    "recipients list cannot be empty",
		},
		{
			name: "mismatched recipients and amounts length",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{validRecipient1, validRecipient2},
				Amounts:    []math.Int{math.NewInt(100)}, // Only 1 amount for 2 recipients
			},
			expectErr: true,
			errMsg:    "amounts list must match recipients list length",
		},
		{
			name: "invalid recipient address",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{"invalid"},
				Amounts:    []math.Int{math.NewInt(100)},
			},
			expectErr: true,
			errMsg:    "invalid recipient address",
		},
		{
			name: "duplicate recipients",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{validRecipient1, validRecipient1},
				Amounts:    []math.Int{math.NewInt(100), math.NewInt(200)},
			},
			expectErr: true,
			errMsg:    "duplicate recipient address",
		},
		{
			name: "zero amount",
			msg: &types.MsgDistributeRewards{
				Sender:     validSender,
				PoolId:     "pool-1",
				Recipients: []string{validRecipient1},
				Amounts:    []math.Int{math.ZeroInt()},
			},
			expectErr: true,
			errMsg:    "amount at index 0 must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetSigners(t *testing.T) {
	validAddress := "cosmos1syavy2npfyt9tcncdtsdzf7kny9lh777pahuux"
	expectedSigners := []sdk.AccAddress{sdk.MustAccAddressFromBech32(validAddress)}

	tests := []struct {
		name string
		msg  interface {
			GetSigners() []sdk.AccAddress
		}
		signers []sdk.AccAddress
	}{
		{
			name:    "MsgUpdateParams",
			msg:     &types.MsgUpdateParams{Authority: validAddress},
			signers: expectedSigners,
		},
		{
			name:    "MsgRegisterMerchant",
			msg:     &types.MsgRegisterMerchant{Sender: validAddress},
			signers: expectedSigners,
		},
		{
			name:    "MsgCreatePayment",
			msg:     &types.MsgCreatePayment{Sender: validAddress},
			signers: expectedSigners,
		},
		{
			name:    "MsgCreateVault",
			msg:     &types.MsgCreateVault{Sender: validAddress},
			signers: expectedSigners,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signers := tt.msg.GetSigners()
			require.Equal(t, tt.signers, signers)
		})
	}
}
