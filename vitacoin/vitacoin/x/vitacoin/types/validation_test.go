package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

const (
	validAddress1 = "vita1hj5fveer5cjtn4wd6wstzugjfdxzl0xp8ws9ct"
	validAddress2 = "vita12luku6uxehhak02py4rcz65zu6swh7wjsrw9wq"
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
				StakeAmount:  math.NewInt(1000000),
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgRegisterMerchant{
				Sender:       invalidAddress,
				BusinessName: "Valid Business",
				StakeAmount:  math.NewInt(1000000),
			},
			expectError: true,
		},
		{
			name: "empty business name",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "",
				StakeAmount:  math.NewInt(1000000),
			},
			expectError: true,
		},
		{
			name: "business name too short",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "AB",
				StakeAmount:  math.NewInt(1000000),
			},
			expectError: true,
		},
		{
			name: "business name too long",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: string(make([]byte, 101)), // 101 characters
				StakeAmount:  math.NewInt(1000000),
			},
			expectError: true,
		},
		{
			name: "business name with control characters",
			msg: types.MsgRegisterMerchant{
				Sender:       validAddress1,
				BusinessName: "Invalid\x00Business",
				StakeAmount:  math.NewInt(1000000),
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
				Amount:    math.NewInt(1e18), // 1 VITA
				Memo:      "Test payment",
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgCreatePayment{
				Sender:    invalidAddress,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1e18),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "invalid recipient address",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: invalidAddress,
				Amount:    math.NewInt(1e18),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "sender and recipient same",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress1,
				Amount:    math.NewInt(1e18),
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
				Amount:    math.NewInt(1e11), // Below minimum
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "amount too large",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1000000000000000000000000000000).Add(math.NewInt(1)),
				Memo:      "Test payment",
			},
			expectError: true,
		},
		{
			name: "memo too long",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1e18),
				Memo:      string(make([]byte, 257)), // 257 characters
			},
			expectError: true,
		},
		{
			name: "memo with invalid control characters",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1e18),
				Memo:      "Invalid\x00memo",
			},
			expectError: true,
		},
		{
			name: "memo with valid control characters (tab, newline, carriage return)",
			msg: types.MsgCreatePayment{
				Sender:    validAddress1,
				MerchantAddress: validAddress2,
				Amount:    math.NewInt(1e18),
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
				Amount:       math.NewInt(1e18), // 1 VITA
				LockDuration: 1000,
			},
			expectError: false,
		},
		{
			name: "invalid owner address",
			msg: types.MsgCreateVault{
				Sender:        invalidAddress,
				Amount:       math.NewInt(1e18),
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
				Amount:       math.NewInt(1000000000000000000000000000000).Add(math.NewInt(1)),
				LockDuration: 1000,
			},
			expectError: true,
		},
		{
			name: "zero unlock height",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1e18),
				LockDuration: 0,
			},
			expectError: true,
		},
		{
			name: "unlock height too far in future",
			msg: types.MsgCreateVault{
				Sender:        validAddress1,
				Amount:       math.NewInt(1e18),
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
			name: "valid message with end height",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        "Test Pool",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   1000,
			},
			expectError: false,
		},
		{
			name: "valid message unlimited pool",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        "Unlimited Pool",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   0, // Unlimited
			},
			expectError: false,
		},
		{
			name: "invalid sender address",
			msg: types.MsgCreateRewardPool{
				Sender:      invalidAddress,
				Name:        "Test Pool",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   1000,
			},
			expectError: true,
		},
		{
			name: "empty pool name",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        "",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   1000,
			},
			expectError: true,
		},
		{
			name: "pool name too short",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        "AB",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   1000,
			},
			expectError: true,
		},
		{
			name: "pool name too long",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        string(make([]byte, 51)), // 51 characters
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   1000,
			},
			expectError: true,
		},
		{
			name: "pool duration too long",
			msg: types.MsgCreateRewardPool{
				Sender:      validAddress1,
				Name:        "Test Pool",
				TotalAmount: math.NewInt(1e18),
				StartHeight: 100,
				EndHeight:   10512101, // Beyond max duration
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
		StakeAmount:        math.NewInt(1e18),
		RegistrationHeight: 100,
		Active:             true,
		TotalVolume:        math.NewInt(5e18),
	}
	merchantStr := merchant.String()
	require.Contains(t, merchantStr, "Test Business")
	require.Contains(t, merchantStr, "Gold")
	require.Contains(t, merchantStr, validAddress1)

	// Test Payment String method
	payment := types.Payment{
		Id:               "payment-123",
		FromAddress:      validAddress1,
		MerchantAddress:        validAddress2,
		Amount:           math.NewInt(1e18),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   100,
		CompletionHeight: 150,
		Fee:              math.NewInt(1e15),
		Memo:             "Test payment",
	}
	paymentStr := payment.String()
	require.Contains(t, paymentStr, "payment-123")
	require.Contains(t, paymentStr, "Completed")
	require.Contains(t, paymentStr, "Test payment")

	// Test Vault String method
	vault := types.Vault{
		Id:               "vault-123",
		Sender:            validAddress1,
		Amount:           math.NewInt(1e18),
		LockDuration:     1000,
		CreationHeight:   100,
		LockDuration:     1100,
		RewardMultiplier: math.LegacyNewDecWithPrec(15, 1), // 1.5
		Withdrawn:        false,
	}
	vaultStr := vault.String()
	require.Contains(t, vaultStr, "vault-123")
	require.Contains(t, vaultStr, validAddress1)
	require.Contains(t, vaultStr, "1000 blocks")

	// Test RewardPool String method
	pool := types.RewardPool{
		Id:                 "pool-123",
		Name:               "Test Pool",
		MerchantAddress:    validAddress1,
		TotalRewards:       math.NewInt(1e18),
		DistributedRewards: math.NewInt(5e17), // 50% distributed
		StartHeight:        100,
		EndHeight:          1000,
		Active:             true,
	}
	poolStr := pool.String()
	require.Contains(t, poolStr, "Test Pool")
	require.Contains(t, poolStr, "50.00%")
	require.Contains(t, poolStr, "1000")
}

func TestMsgStringMethods(t *testing.T) {
	// Test MsgRegisterMerchant String method
	msg := &types.MsgRegisterMerchant{
		Sender:       validAddress1,
		BusinessName: "Test Business",
		StakeAmount:  math.NewInt(1e18),
	}
	msgStr := msg.String()
	require.Contains(t, msgStr, "MsgRegisterMerchant")
	require.Contains(t, msgStr, "Test Business")
	require.Contains(t, msgStr, validAddress1)

	// Test MsgCreatePayment String method
	paymentMsg := &types.MsgCreatePayment{
		Sender:    validAddress1,
		MerchantAddress: validAddress2,
		Amount:    math.NewInt(1e18),
		Memo:      "Test payment",
	}
	paymentMsgStr := paymentMsg.String()
	require.Contains(t, paymentMsgStr, "MsgCreatePayment")
	require.Contains(t, paymentMsgStr, "Test payment")
	require.Contains(t, paymentMsgStr, validAddress2)

	// Test MsgDistributeRewards String method
	distributeMsg := &types.MsgDistributeRewards{
		Sender:   validAddress1,
		PoolName: "Test Pool",
		Recipients: []types.RewardRecipient{
			{Address: validAddress1, Amount: math.NewInt(1e17)},
			{Address: validAddress2, Amount: math.NewInt(2e17)},
		},
	}
	distributeMsgStr := distributeMsg.String()
	require.Contains(t, distributeMsgStr, "MsgDistributeRewards")
	require.Contains(t, distributeMsgStr, "Test Pool")
	require.Contains(t, distributeMsgStr, "2")
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
	require.Contains(t, paramsStr, "MinGasPrice")
	require.Contains(t, paramsStr, "TransactionFeePercent")
}