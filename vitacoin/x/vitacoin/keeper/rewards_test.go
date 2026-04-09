package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testDelegatorAddr builds a deterministic vita bech32 address for tests.
func testDelegatorAddr(idx byte) string {
	priv := make([]byte, 20)
	priv[0] = idx
	return sdk.AccAddress(priv).String()
}

// TestAccrueDelegatorReward: accrue 100 VITA reward, verify GetPendingRewards returns 100.
func (suite *KeeperTestSuite) TestAccrueDelegatorReward() {
	delegator := testDelegatorAddr(0x10)
	amount := math.NewInt(100)

	err := suite.keeper.AccrueDelegatorReward(suite.ctx, delegator, amount, "uvita")
	suite.Require().NoError(err)

	pending, denom, err := suite.keeper.GetPendingRewards(suite.ctx, delegator)
	suite.Require().NoError(err)
	suite.Require().Equal(amount, pending, fmt.Sprintf("expected %s pending rewards", amount))
	suite.Require().Equal("uvita", denom)
}

// TestAccrueDelegatorReward_Accumulates: accrue 100 then 50, verify total = 150.
func (suite *KeeperTestSuite) TestAccrueDelegatorReward_Accumulates() {
	delegator := testDelegatorAddr(0x11)
	first := math.NewInt(100)
	second := math.NewInt(50)

	err := suite.keeper.AccrueDelegatorReward(suite.ctx, delegator, first, "uvita")
	suite.Require().NoError(err)
	err = suite.keeper.AccrueDelegatorReward(suite.ctx, delegator, second, "uvita")
	suite.Require().NoError(err)

	pending, _, err := suite.keeper.GetPendingRewards(suite.ctx, delegator)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(150), pending)
}

// TestClaimDelegatorRewards: accrue reward, claim it, verify KV cleared.
func (suite *KeeperTestSuite) TestClaimDelegatorRewards() {
	delegator := testDelegatorAddr(0x12)
	amount := math.NewInt(200)

	err := suite.keeper.AccrueDelegatorReward(suite.ctx, delegator, amount, "uvita")
	suite.Require().NoError(err)

	claimed, err := suite.keeper.ClaimDelegatorRewards(suite.ctx, delegator)
	suite.Require().NoError(err)
	suite.Require().Len(claimed, 1)
	suite.Require().Equal(amount, claimed[0].Amount)

	// After claiming, KV should be cleared — pending should be zero
	pending, _, err := suite.keeper.GetPendingRewards(suite.ctx, delegator)
	suite.Require().NoError(err)
	suite.Require().True(pending.IsZero(), "pending rewards should be zero after claim")
}
