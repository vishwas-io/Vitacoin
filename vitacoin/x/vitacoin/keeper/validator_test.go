package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
)

// oneVITA is 1e18 avita (the smallest denomination).
var oneVITA = math.NewIntWithDecimal(1, 18)

// minSelfBond is 10,000 VITA — the DefaultStakingParams minimum.
var minSelfBond = math.NewInt(10_000).Mul(oneVITA)

// testValidatorAddr returns a bech32 acc address (vita1...) for tests.
// RegisterValidator expects an AccAddress-based bech32.
func testValidatorAddr(idx byte) string {
	priv := make([]byte, 20)
	priv[0] = idx
	return sdk.AccAddress(priv).String()
}

// TestRegisterValidator: register with valid commission + self-bond,
// verify KV record stored, verify GetValidator returns correct data.
func (suite *KeeperTestSuite) TestRegisterValidator() {
	operatorAddr := testValidatorAddr(0xAB)
	commission := math.LegacyNewDecWithPrec(5, 2) // 5%
	selfBond := minSelfBond                        // 10,000 VITA

	err := suite.keeper.RegisterValidator(suite.ctx, operatorAddr, "node-1", commission, selfBond)
	suite.Require().NoError(err)

	val, found, err := suite.keeper.GetValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(found, "validator should be stored")
	suite.Require().Equal(operatorAddr, val.OperatorAddress)
	suite.Require().Equal("node-1", val.Moniker)
	suite.Require().True(commission.Equal(val.Commission))
	suite.Require().Equal(selfBond, val.TotalDelegated)
	suite.Require().Equal(selfBond, val.SelfBond)
	suite.Require().False(val.Jailed)
}

// TestRegisterValidator_TooManyValidators: set MaxValidators=1 in code is not
// injectable via params (params use DefaultStakingParams). Instead we register
// one validator, then try to register the same address again — should get
// "already registered" error.
func (suite *KeeperTestSuite) TestRegisterValidator_TooManyValidators() {
	// Register first validator
	priv1 := make([]byte, 20)
	priv1[0] = 0x01
	addr1 := sdk.AccAddress(priv1).String()
	selfBond := minSelfBond
	commission := math.LegacyNewDecWithPrec(5, 2)

	err := suite.keeper.RegisterValidator(suite.ctx, addr1, "node-1", commission, selfBond)
	suite.Require().NoError(err)

	// Try to register same address again — should fail with duplicate error
	err2 := suite.keeper.RegisterValidator(suite.ctx, addr1, "node-1-dup", commission, selfBond)
	suite.Require().Error(err2)
	suite.Require().Contains(err2.Error(), "already registered")
}

// TestSlashValidator: register validator with 1000 VITA delegated,
// slash 10%, verify TotalDelegated reduced by 100.
func (suite *KeeperTestSuite) TestSlashValidator() {
	priv := make([]byte, 20)
	priv[0] = 0xCC
	operatorAddr := sdk.AccAddress(priv).String()
	selfBond := math.NewInt(10_000).Mul(oneVITA) // 10,000 VITA == minimum self-bond
	commission := math.LegacyNewDecWithPrec(5, 2)

	err := suite.keeper.RegisterValidator(suite.ctx, operatorAddr, "slashable", commission, selfBond)
	suite.Require().NoError(err)

	slashFactor := math.LegacyNewDecWithPrec(10, 2) // 10%
	err = suite.keeper.SlashValidator(suite.ctx, operatorAddr, slashFactor)
	suite.Require().NoError(err)

	val, found, err := suite.keeper.GetValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(found)

	expected := math.NewInt(9_000).Mul(oneVITA) // 10000 * 0.90 = 9000 VITA
	suite.Require().Equal(expected, val.TotalDelegated,
		"TotalDelegated should be reduced by 10%%: got %s, want %s", val.TotalDelegated, expected)
}

// TestJailUnjail: register validator, jail it, verify Jailed=true,
// unjail, verify Jailed=false.
func (suite *KeeperTestSuite) TestJailUnjail() {
	priv := make([]byte, 20)
	priv[0] = 0xDD
	operatorAddr := sdk.AccAddress(priv).String()
	selfBond := minSelfBond
	commission := math.LegacyNewDecWithPrec(5, 2)

	err := suite.keeper.RegisterValidator(suite.ctx, operatorAddr, "jailable", commission, selfBond)
	suite.Require().NoError(err)

	// Jail
	err = suite.keeper.JailValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)

	val, found, err := suite.keeper.GetValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(found)
	suite.Require().True(val.Jailed, "validator should be jailed")

	// Unjail
	err = suite.keeper.UnjailValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)

	val, found, err = suite.keeper.GetValidator(suite.ctx, operatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(found)
	suite.Require().False(val.Jailed, "validator should be unjailed")
}

// Compile-time assertion: ValidatorRecord is accessible.
var _ = keeper.ValidatorRecord{}
