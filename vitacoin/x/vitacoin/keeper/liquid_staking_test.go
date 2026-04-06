package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestGetStVITASupply_Initial: verify returns ZeroInt before any minting.
func (suite *KeeperTestSuite) TestGetStVITASupply_Initial() {
	supply := suite.keeper.GetStVITASupply(suite.ctx)
	suite.Require().True(supply.IsZero(), "initial stVITA supply should be zero")
}

// TestMintBurnStVITA: mint 1000 stVITA, verify supply=1000; burn 500, verify supply=500.
func (suite *KeeperTestSuite) TestMintBurnStVITA() {
	addr := sdk.AccAddress(make([]byte, 20))
	mintAmount := math.NewInt(1000)

	err := suite.keeper.MintStVITA(suite.ctx, addr, mintAmount)
	suite.Require().NoError(err)

	supply := suite.keeper.GetStVITASupply(suite.ctx)
	suite.Require().Equal(mintAmount, supply, "supply should be 1000 after minting")

	burnAmount := math.NewInt(500)
	err = suite.keeper.BurnStVITA(suite.ctx, addr, burnAmount)
	suite.Require().NoError(err)

	supply = suite.keeper.GetStVITASupply(suite.ctx)
	suite.Require().Equal(math.NewInt(500), supply, "supply should be 500 after burning 500")
}

// TestGetStVITAExchangeRate_Default: verify returns Dec(1) when supply=0.
func (suite *KeeperTestSuite) TestGetStVITAExchangeRate_Default() {
	rate := suite.keeper.GetStVITAExchangeRate(suite.ctx)
	suite.Require().True(math.LegacyOneDec().Equal(rate),
		"default exchange rate should be 1, got %s", rate)
}
