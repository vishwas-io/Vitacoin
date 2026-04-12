package types_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

func init() {
	// Set bech32 prefixes
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")
	config.SetBech32PrefixForValidator("vitavaloper", "vitavaloperpub")
	config.SetBech32PrefixForConsensusNode("vitavalcons", "vitavalconspub")
}

const (
	validAddress1 = "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f" // testmerchant1
	validAddress2 = "vita1x0xrzpm2h89smwsapxdhtualwh8w0968vp48k4" // testmerchant2
	invalidAddress = "invalid_address"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectError bool
	}{
		{
			name: "valid message",
			msg: types.MsgUpdateParams{
				Authority: validAddress1,
				Params:    types.DefaultParams(),
			},
			expectError: false,
		},
		{
			name: "invalid authority address",
			msg: types.MsgUpdateParams{
				Authority: invalidAddress,
				Params:    types.DefaultParams(),
			},
			expectError: true,
		},
		{
			name: "invalid params",
			msg: types.MsgUpdateParams{
				Authority: validAddress1,
				Params: types.Params{
					MinGasPrice:           math.LegacyNewDec(-1), // Invalid negative value
					TransactionFeePercent: math.LegacyNewDec(50),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRegisterMerchant_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgRegisterMerchant
		expectError bool
	}{
		{
			name: "valid message",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Valid Business",
				StakeAmount:  math.NewInt(10000000000000), // 10000 VITA (minimum)
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgRegisterMerchant{
				Sender:       invalidAddress,
				BusinessName: "Valid Business",
				StakeAmount:  math.NewInt(10000000000000),
			},
			expectError: true,
		},
		{
			name: "empty business name",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "",
				StakeAmount:  math.NewInt(10000000000000),
			},
			expectError: true,
		},
		{
			name: "business name too short",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "AB",
				StakeAmount:  math.NewInt(10000000000000),
			},
			expectError: true,
		},
		{
			name: "business name too long",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: string(make([]byte, 101)), // 101 characters
				StakeAmount:  math.NewInt(10000000000000),
			},
			expectError: true,
		},
		{
			name: "business name with control characters",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Invalid\x00Business",
				StakeAmount:  math.NewInt(10000000000000),
			},
			expectError: true,
		},
		{
			name: "zero stake amount",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Valid Business",
				StakeAmount:  math.ZeroInt(),
			},
			expectError: true,
		},
		{
			name: "negative stake amount",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Valid Business",
				StakeAmount:  math.NewInt(-1000),
			},
			expectError: true,
		},
		{
			name: "stake amount too large",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Valid Business",
				StakeAmount:  math.NewInt(1000000000).Mul(math.NewInt(1000000000)).Mul(math.NewInt(1000)).Add(math.NewInt(1)),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreatePayment_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgCreatePayment
		expectError bool
	}{
		{
			name: "valid message",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000), // 1 VITA
				Memo:      "Test payment",
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgCreatePayment{
				Sender:    invalidAddress,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "invalid recipient address",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: invalidAddress,
				Amount:    math.NewInt(1000000),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "sender and recipient same",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress1,
				Amount:    math.NewInt(1000000),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "zero amount",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.ZeroInt(),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "amount too small",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(999), // Below MinPaymentAmount (1000)
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "amount too large",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewIntFromBigInt(new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "memo too long",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000),
				Memo:      string(make([]byte, 257)), // 257 characters
			},
			expectError: true,
		},
		{
			name: "memo with invalid control characters",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000),
				Memo:      "Invalid\x00memo",
			},
			expectError: true,
		},
		{
			name: "memo with valid control characters (tab, newline, carriage return)",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000),
				Memo:      "Valid\t\n\rmemo",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateVault_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgCreateVault
		expectError bool
	}{
		{
			name: "valid message",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1000000), // 1 VITA
				LockDuration: 1000,
			},
			expectError: false,
		},
		{
			name: "invalid owner address",
			msg: types.MsgCreateVault{
				Sender:        invalidAddress,
				Amount:       math.NewInt(1000000),
				LockDuration: 1000,
			},
			expectError: true,
		},
		{
			name: "zero amount",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.ZeroInt(),
				LockDuration: 1000,
			},
			expectError: true,
		},
		{
			name: "amount below minimum",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1e17), // Below 1 VITA minimum
				LockDuration: 1000,
			},
			expectError: true,
		},
		{
			name: "amount too large",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewIntFromBigInt(new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})),
				LockDuration: 1000,
			},
			expectError: true,
		},
		{
			name: "zero unlock height",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1000000),
				LockDuration: 0,
			},
			expectError: true,
		},
		{
			name: "unlock height too far in future",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1000000),
				LockDuration: 100000001, // Beyond max
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateRewardPool_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         types.MsgCreateRewardPool
		expectError bool
	}{
		{
			name: "valid message with duration",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.NewInt(1000000), // 1e15 (0.001 VITA minimum)
				DurationBlocks: 1000,
			},
			expectError: false,
		},
		{
			name: "valid message unlimited pool",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.NewInt(1000000), // 1e15 (0.001 VITA minimum)
				DurationBlocks: 0, // Unlimited
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgCreateRewardPool{
				Sender:         invalidAddress,
				TotalRewards:   math.NewInt(1000000000000),
				DurationBlocks: 1000,
			},
			expectError: true,
		},
		{
			name: "zero total rewards",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.ZeroInt(),
				DurationBlocks: 1000,
			},
			expectError: true,
		},
		{
			name: "negative total rewards",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.NewInt(-1000),
				DurationBlocks: 1000,
			},
			expectError: true,
		},
		{
			name: "rewards too large",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.NewIntFromBigInt(new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})),
				DurationBlocks: 1000,
			},
			expectError: true,
		},
		{
			name: "duration blocks too long",
			msg: types.MsgCreateRewardPool{
				Sender:         validAddress1,
				TotalRewards:   math.NewInt(1000000),
				DurationBlocks: 100000001, // Beyond max duration
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEntityStringMethods(t *testing.T) {
	// Test Merchant String method
	merchant := types.Merchant{
		Address:            validAddress1,
		BusinessName:       "Test Business",
		Tier:               types.MerchantTierGold,
		StakeAmount:        math.NewInt(1000000),
		RegistrationHeight: 100,
		IsActive:           true,
		TotalVolume:        math.NewInt(5e18),
	}
	merchantStr := merchant.String()
	require.Contains(t, merchantStr, "Test Business")
	require.Contains(t, merchantStr, "MERCHANT_TIER_GOLD")
	require.Contains(t, merchantStr, validAddress1)

	// Test Payment String method
	payment := types.Payment{
		Id:               "payment-123",
		FromAddress:      validAddress1,
		ToAddress:        validAddress2,
		Amount:           math.NewInt(1000000),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   100,
		CompletionHeight: 150,
		Memo:             "Test payment",
	}
	paymentStr := payment.String()
	require.Contains(t, paymentStr, "payment-123")
	require.Contains(t, paymentStr, "PAYMENT_STATUS_COMPLETED")
	require.Contains(t, paymentStr, "Test payment")

	// Test Vault String method
	vault := types.Vault{
		Id:               "vault-123",
		Owner:            validAddress1,
		Amount:           math.NewInt(1000000),
		LockDuration:     1000,
		CreationHeight:   100,
		UnlockHeight:     1100,
		RewardMultiplier: math.LegacyNewDecWithPrec(15, 1), // 1.5
	}
	vaultStr := vault.String()
	require.Contains(t, vaultStr, "vault-123")
	require.Contains(t, vaultStr, validAddress1)
	require.Contains(t, vaultStr, "1000")

	// Test RewardPool String method
	pool := types.RewardPool{
		Id:                 "pool-123",
		MerchantAddress:    validAddress1,
		TotalRewards:       math.NewInt(1000000),
		DistributedRewards: math.NewInt(5e17), // 50% distributed
		StartHeight:        100,
		EndHeight:          1000,
		IsActive:           true,
	}
	poolStr := pool.String()
	require.Contains(t, poolStr, "pool-123")
	require.Contains(t, poolStr, validAddress1)
	require.Contains(t, poolStr, "1000")
}

func TestMsgStringMethods(t *testing.T) {
	// Test MsgRegisterMerchant String method
	msg := &types.MsgRegisterMerchant{
		Sender:       validAddress1,
		BusinessName: "Test Business",
		StakeAmount:  math.NewInt(1000000),
	}
	msgStr := msg.String()
	require.Contains(t, msgStr, "sender")
	require.Contains(t, msgStr, "Test Business")
	require.Contains(t, msgStr, validAddress1)

	// Test MsgCreatePayment String method
	paymentMsg := &types.MsgCreatePayment{
		Sender:    validAddress1,
		MerchantAddress: validAddress2,
		Amount:    math.NewInt(1000000),
		Memo:      "Test payment",
	}
	paymentMsgStr := paymentMsg.String()
	require.Contains(t, paymentMsgStr, "sender")
	require.Contains(t, paymentMsgStr, "Test payment")
	require.Contains(t, paymentMsgStr, validAddress2)

	// Test MsgDistributeRewards String method
	distributeMsg := &types.MsgDistributeRewards{
		Sender:     validAddress1,
		PoolId:     "pool-123",
		Recipients: []string{validAddress1, validAddress2},
		Amounts:    []math.Int{math.NewInt(1e17), math.NewInt(2e17)},
	}
	distributeMsgStr := distributeMsg.String()
	require.Contains(t, distributeMsgStr, "sender")
	require.Contains(t, distributeMsgStr, "pool-123")
	require.Contains(t, distributeMsgStr, validAddress1)
}

func TestDefaultParamsValidation(t *testing.T) {
	params := types.DefaultParams()
	
	// Test that default values are reasonable
	require.False(t, params.MinGasPrice.IsNegative())
	require.True(t, params.TransactionFeePercent.LTE(math.LegacyNewDec(100)))
	require.True(t, params.MerchantFeeDiscount.LTE(math.LegacyNewDec(100)))
	require.False(t, params.MerchantRegistrationFee.IsNegative())
	require.True(t, params.PaymentTimeoutBlocks > 0)
	
	// Test that params validate correctly
	require.NoError(t, params.Validate())
	
	// Test String method
	paramsStr := params.String()
	require.Contains(t, paramsStr, "Min Gas Price")
	require.Contains(t, paramsStr, "Transaction Fee Percent")
}