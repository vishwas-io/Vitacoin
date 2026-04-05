package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func testDelegatorAddr() sdk.AccAddress {
	return sdk.AccAddress([]byte("delegator1__________"))
}

func testValidatorAddr() sdk.ValAddress {
	return sdk.ValAddress([]byte("validator1__________"))
}

func testVITACoin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(types.BondDenom, amount)
}

// ---------------------------------------------------------------------------
// MsgDelegateVITA tests
// ---------------------------------------------------------------------------

func TestMsgDelegateVITA_ValidateBasic(t *testing.T) {
	delegator := testDelegatorAddr()
	validator := testValidatorAddr()

	tests := []struct {
		name    string
		msg     *types.MsgDelegateVITA
		wantErr bool
	}{
		{
			name:    "valid delegation",
			msg:     types.NewMsgDelegateVITA(delegator, validator, testVITACoin(1_000_000_000_000_000_000)),
			wantErr: false,
		},
		{
			name:    "empty delegator",
			msg:     types.NewMsgDelegateVITA(sdk.AccAddress{}, validator, testVITACoin(100)),
			wantErr: true,
		},
		{
			name:    "empty validator",
			msg:     types.NewMsgDelegateVITA(delegator, sdk.ValAddress{}, testVITACoin(100)),
			wantErr: true,
		},
		{
			name:    "zero amount",
			msg:     types.NewMsgDelegateVITA(delegator, validator, sdk.NewInt64Coin(types.BondDenom, 0)),
			wantErr: true,
		},
		{
			name:    "wrong denom",
			msg:     types.NewMsgDelegateVITA(delegator, validator, sdk.NewInt64Coin("uatom", 100)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgDelegateVITA_GetSigners(t *testing.T) {
	delegator := testDelegatorAddr()
	validator := testValidatorAddr()
	msg := types.NewMsgDelegateVITA(delegator, validator, testVITACoin(1000))
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, delegator, signers[0])
}

// ---------------------------------------------------------------------------
// MsgUndelegateVITA tests
// ---------------------------------------------------------------------------

func TestMsgUndelegateVITA_ValidateBasic(t *testing.T) {
	delegator := testDelegatorAddr()
	validator := testValidatorAddr()

	tests := []struct {
		name    string
		msg     *types.MsgUndelegateVITA
		wantErr bool
	}{
		{
			name:    "valid undelegate",
			msg:     types.NewMsgUndelegateVITA(delegator, validator, testVITACoin(500)),
			wantErr: false,
		},
		{
			name:    "empty delegator",
			msg:     types.NewMsgUndelegateVITA(sdk.AccAddress{}, validator, testVITACoin(100)),
			wantErr: true,
		},
		{
			name:    "empty validator",
			msg:     types.NewMsgUndelegateVITA(delegator, sdk.ValAddress{}, testVITACoin(100)),
			wantErr: true,
		},
		{
			name:    "wrong denom",
			msg:     types.NewMsgUndelegateVITA(delegator, validator, sdk.NewInt64Coin("uatom", 100)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MsgClaimStakingRewards tests
// ---------------------------------------------------------------------------

func TestMsgClaimStakingRewards_ValidateBasic(t *testing.T) {
	delegator := testDelegatorAddr()

	require.NoError(t, types.NewMsgClaimStakingRewards(delegator).ValidateBasic())
	require.Error(t, types.NewMsgClaimStakingRewards(sdk.AccAddress{}).ValidateBasic())
}

func TestMsgClaimStakingRewards_GetSigners(t *testing.T) {
	delegator := testDelegatorAddr()
	msg := types.NewMsgClaimStakingRewards(delegator)
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, delegator, signers[0])
}

// ---------------------------------------------------------------------------
// MsgCreateValidator tests
// ---------------------------------------------------------------------------

func validDescription() types.ValidatorDescription {
	return types.ValidatorDescription{
		Moniker:         "TestValidator",
		Identity:        "ABCD1234",
		Website:         "https://validator.example.com",
		SecurityContact: "security@example.com",
		Details:         "A reliable VitaCoin validator.",
	}
}

func validCommission() types.CommissionRates {
	return types.CommissionRates{
		Rate:          math.LegacyNewDecWithPrec(5, 2),  // 5%
		MaxRate:       math.LegacyNewDecWithPrec(20, 2), // 20%
		MaxChangeRate: math.LegacyNewDecWithPrec(1, 2),  // 1% per day
	}
}

func TestMsgCreateValidator_ValidateBasic(t *testing.T) {
	operator := testValidatorAddr()
	oneVITA := int64(1_000_000_000_000_000_000)

	tests := []struct {
		name    string
		msg     *types.MsgCreateValidator
		wantErr bool
	}{
		{
			name: "valid create validator",
			msg: types.NewMsgCreateValidator(
				operator,
				validDescription(),
				validCommission(),
				math.NewInt(10_000*oneVITA),
				"ABC123pubkey==",
				sdk.NewInt64Coin(types.BondDenom, 10_000*oneVITA),
			),
			wantErr: false,
		},
		{
			name: "empty operator",
			msg: types.NewMsgCreateValidator(
				sdk.ValAddress{},
				validDescription(),
				validCommission(),
				math.NewInt(oneVITA),
				"ABC123",
				sdk.NewInt64Coin(types.BondDenom, oneVITA),
			),
			wantErr: true,
		},
		{
			name: "empty moniker",
			msg: types.NewMsgCreateValidator(
				operator,
				types.ValidatorDescription{Moniker: ""},
				validCommission(),
				math.NewInt(oneVITA),
				"ABC123",
				sdk.NewInt64Coin(types.BondDenom, oneVITA),
			),
			wantErr: true,
		},
		{
			name: "initial delegation < min self delegation",
			msg: types.NewMsgCreateValidator(
				operator,
				validDescription(),
				validCommission(),
				math.NewInt(1000*oneVITA),
				"ABC123",
				sdk.NewInt64Coin(types.BondDenom, oneVITA), // only 1 VITA, need 1000
			),
			wantErr: true,
		},
		{
			name: "wrong denom",
			msg: types.NewMsgCreateValidator(
				operator,
				validDescription(),
				validCommission(),
				math.NewInt(oneVITA),
				"ABC123",
				sdk.NewInt64Coin("uatom", oneVITA),
			),
			wantErr: true,
		},
		{
			name: "empty pubkey",
			msg: types.NewMsgCreateValidator(
				operator,
				validDescription(),
				validCommission(),
				math.NewInt(oneVITA),
				"",
				sdk.NewInt64Coin(types.BondDenom, oneVITA),
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidatorDescription tests
// ---------------------------------------------------------------------------

func TestValidatorDescription_Validate(t *testing.T) {
	require.NoError(t, validDescription().Validate())
	require.Error(t, types.ValidatorDescription{Moniker: ""}.Validate())
	require.Error(t, types.ValidatorDescription{Moniker: string(make([]byte, 71))}.Validate())
}

// ---------------------------------------------------------------------------
// CommissionRates tests
// ---------------------------------------------------------------------------

func TestCommissionRates_Validate(t *testing.T) {
	require.NoError(t, validCommission().Validate())

	// Rate > MaxRate
	require.Error(t, types.CommissionRates{
		Rate:          math.LegacyNewDecWithPrec(30, 2),
		MaxRate:       math.LegacyNewDecWithPrec(20, 2),
		MaxChangeRate: math.LegacyNewDecWithPrec(1, 2),
	}.Validate())

	// MaxRate > 1
	require.Error(t, types.CommissionRates{
		Rate:          math.LegacyNewDecWithPrec(5, 2),
		MaxRate:       math.LegacyNewDecWithPrec(110, 2),
		MaxChangeRate: math.LegacyNewDecWithPrec(1, 2),
	}.Validate())
}

// ---------------------------------------------------------------------------
// StakingParams tests
// ---------------------------------------------------------------------------

func TestDefaultStakingParams_Valid(t *testing.T) {
	p := types.DefaultStakingParams()
	require.NoError(t, p.Validate())
}

func TestStakingParams_Validate(t *testing.T) {
	p := types.DefaultStakingParams()

	// Zero unbonding time
	bad := p
	bad.UnbondingTime = 0
	require.Error(t, bad.Validate())

	// Zero max validators
	bad = p
	bad.MaxValidators = 0
	require.Error(t, bad.Validate())

	// Negative reward rate
	bad = p
	bad.StakingRewardRate = math.LegacyNewDecWithPrec(-1, 2)
	require.Error(t, bad.Validate())
}

func TestStakingParams_Defaults(t *testing.T) {
	p := types.DefaultStakingParams()
	require.Equal(t, 21*24*time.Hour, p.UnbondingTime)
	require.Equal(t, uint32(100), p.MaxValidators)
	require.Equal(t, "0.100000000000000000", p.StakingRewardRate.String())
}
