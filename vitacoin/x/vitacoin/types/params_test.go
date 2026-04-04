package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.NotNil(t, params)
	require.NoError(t, params.Validate())

	// Verify default values
	require.Equal(t, math.LegacyNewDecWithPrec(1, 3), params.MinGasPrice)
	require.Equal(t, math.LegacyNewDecWithPrec(1, 3), params.TransactionFeePercent) // 0.1%
	require.Equal(t, math.LegacyZeroDec(), params.MerchantFeeDiscount) // 50%
	require.True(t, params.EnableMerchantLoyalty)
	require.True(t, params.EnableInstantSettlement)
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name      string
		params    types.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default params",
			params:    types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "negative min gas price",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDec(-1),
				TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:   math.LegacyNewDec(50),
			},
			expectErr: true,
			errMsg:    "min gas price must be non-negative",
		},
		{
			name: "negative transaction fee percent",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDec(-1),
				MerchantFeeDiscount:   math.LegacyNewDec(50),
			},
			expectErr: true,
			errMsg:    "transaction fee percent must be between 0 and 100",
		},
		{
			name: "transaction fee percent over 100",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDec(101),
				MerchantFeeDiscount:   math.LegacyNewDec(50),
			},
			expectErr: true,
			errMsg:    "transaction fee percent must be between 0 and 100",
		},
		{
			name: "negative merchant fee discount",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:   math.LegacyNewDec(-1),
			},
			expectErr: true,
			errMsg:    "merchant fee discount must be between 0 and 100",
		},
		{
			name: "merchant fee discount over 100",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:   math.LegacyNewDec(101),
			},
			expectErr: true,
			errMsg:    "merchant fee discount must be between 0 and 100",
		},
		{
			name: "negative max transaction amount",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:   math.LegacyNewDec(50),
				MaxTransactionAmount:  math.NewInt(-1),
			},
			expectErr: true,
			errMsg:    "max transaction amount must be non-negative",
		},
		{
			name: "zero payment timeout blocks",
			params: types.Params{
				MinGasPrice:           math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:   math.LegacyNewDec(50),
				MaxTransactionAmount:  math.ZeroInt(),
				PaymentTimeoutBlocks:  0,
			},
			expectErr: true,
			errMsg:    "payment timeout blocks must be greater than 0",
		},
		{
			name: "negative merchant registration fee",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(-1),
			},
			expectErr: true,
			errMsg:    "merchant registration fee must be non-negative",
		},
		{
			name: "negative loyalty reward percent",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(1e18),
				LoyaltyRewardPercent:    math.LegacyNewDec(-1),
			},
			expectErr: true,
			errMsg:    "loyalty reward percent must be between 0 and 100",
		},
		{
			name: "loyalty reward percent over 100",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(1e18),
				LoyaltyRewardPercent:    math.LegacyNewDec(101),
			},
			expectErr: true,
			errMsg:    "loyalty reward percent must be between 0 and 100",
		},
		{
			name: "negative min merchant stake",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(1e18),
				LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(1, 0),
				MinMerchantStake:        math.NewInt(-1),
			},
			expectErr: true,
			errMsg:    "min merchant stake must be non-negative",
		},
		{
			name: "negative fee burn percent",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(1e18),
				LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(1, 0),
				MinMerchantStake:        math.NewIntFromUint64(10e18),
				FeeBurnPercent:          math.LegacyNewDec(-1),
			},
			expectErr: true,
			errMsg:    "fee burn percent must be between 0 and 100",
		},
		{
			name: "fee burn percent over 100",
			params: types.Params{
				MinGasPrice:             math.LegacyNewDecWithPrec(1, 3),
				TransactionFeePercent:   math.LegacyNewDecWithPrec(5, 1),
				MerchantFeeDiscount:     math.LegacyNewDec(50),
				MaxTransactionAmount:    math.ZeroInt(),
				PaymentTimeoutBlocks:    100,
				MerchantRegistrationFee: math.NewInt(1e18),
				LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(1, 0),
				MinMerchantStake:        math.NewIntFromUint64(10e18),
				FeeBurnPercent:          math.LegacyNewDec(101),
			},
			expectErr: true,
			errMsg:    "fee burn percent must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParamsString(t *testing.T) {
	params := types.DefaultParams()
	str := params.String()

	require.NotEmpty(t, str)
	require.Contains(t, str, "Vitacoin Params")
	require.Contains(t, str, "Min Gas Price")
	require.Contains(t, str, "Transaction Fee Percent")
	require.Contains(t, str, "Merchant Fee Discount")
}

func TestDefaultGenesisState(t *testing.T) {
	genState := types.DefaultGenesisState()

	require.NotNil(t, genState)
	require.NoError(t, genState.Validate())

	// Verify default values
	require.Equal(t, types.DefaultParams(), genState.Params)
	require.Empty(t, genState.MerchantList)
	require.Empty(t, genState.PaymentList)
	require.Empty(t, genState.VaultList)
	require.Empty(t, genState.PoolList)
}

func TestGenesisStateValidate(t *testing.T) {
	tests := []struct {
		name      string
		genState  *types.GenesisState
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default genesis",
			genState:  types.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "invalid params",
			genState: &types.GenesisState{
				Params: types.Params{
					MinGasPrice:           math.LegacyNewDec(-1),
					TransactionFeePercent: math.LegacyNewDecWithPrec(5, 1),
				},
				MerchantList: []types.Merchant{},
				PaymentList:  []types.Payment{},
				VaultList:    []types.Vault{},
				PoolList:     []types.RewardPool{},
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
		{
			name: "duplicate merchant addresses",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				MerchantList: []types.Merchant{
					{
						Address:      "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
						BusinessName: "Test 1",
						StakeAmount:  math.NewInt(1000),
						TotalVolume:  math.NewInt(0),
						Tier:         0,
					},
					{
						Address:      "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
						BusinessName: "Test 2",
						StakeAmount:  math.NewInt(2000),
						TotalVolume:  math.NewInt(0),
						Tier:         0,
					},
				},
				PaymentList: []types.Payment{},
				VaultList:   []types.Vault{},
				PoolList:    []types.RewardPool{},
			},
			expectErr: true,
			errMsg:    "duplicate merchant address",
		},
		{
			name: "duplicate payment IDs",
			genState: &types.GenesisState{
				Params:       types.DefaultParams(),
				MerchantList: []types.Merchant{},
				PaymentList: []types.Payment{
					{
						Id:             "payment-1",
						FromAddress:    "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
						ToAddress:      "vita1x0xrzpm2h89smwsapxdhtualwh8w0968vp48k4",
						Amount:         math.NewInt(1000),
						Status:         types.PaymentStatusPending,
						CreationHeight: 100,
					},
					{
						Id:             "payment-1",
						FromAddress:    "vita1x0xrzpm2h89smwsapxdhtualwh8w0968vp48k4",
						ToAddress:      "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
						Amount:         math.NewInt(2000),
						Status:         types.PaymentStatusPending,
						CreationHeight: 101,
					},
				},
				VaultList: []types.Vault{},
				PoolList:  []types.RewardPool{},
			},
			expectErr: true,
			errMsg:    "duplicate payment ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.genState.Validate()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
