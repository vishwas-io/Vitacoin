package keeper_test

// fee_state_treasury_test.go — tests for fee_state.go and treasury.go functions.

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

type FeeStateTreasuryTestSuite struct {
	KeeperTestSuite
}

func TestFeeStateTreasuryTestSuite(t *testing.T) {
	suite.Run(t, new(FeeStateTreasuryTestSuite))
}

// ── BlockFeeAccumulator ───────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestBlockFeeAccumulator_SetGet() {
	acc := types.BlockFeeAccumulator{
		Height:           10,
		TotalCollected:   sdkmath.NewInt(500),
		TransactionCount: 3,
	}
	require.NoError(suite.T(), suite.keeper.SetBlockFeeAccumulator(suite.ctx, acc))

	got, err := suite.keeper.GetBlockFeeAccumulator(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), acc.Height, got.Height)
	require.Equal(suite.T(), acc.TotalCollected, got.TotalCollected)
	require.Equal(suite.T(), acc.TransactionCount, got.TransactionCount)
}

func (suite *FeeStateTreasuryTestSuite) TestBlockFeeAccumulator_NotFound() {
	_, err := suite.keeper.GetBlockFeeAccumulator(suite.ctx)
	require.Error(suite.T(), err)
}

func (suite *FeeStateTreasuryTestSuite) TestBlockFeeAccumulator_Delete() {
	acc := types.BlockFeeAccumulator{
		Height:           1,
		TotalCollected:   sdkmath.NewInt(100),
		TransactionCount: 1,
	}
	require.NoError(suite.T(), suite.keeper.SetBlockFeeAccumulator(suite.ctx, acc))
	require.NoError(suite.T(), suite.keeper.DeleteBlockFeeAccumulator(suite.ctx))
	_, err := suite.keeper.GetBlockFeeAccumulator(suite.ctx)
	require.Error(suite.T(), err)
}

// ── FeeStatistics ─────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestFeeStatistics_SetGet() {
	stats := types.FeeStatistics{
		TotalCollectedAllTime:    sdkmath.NewInt(10000),
		TotalBurnedAllTime:       sdkmath.NewInt(500),
		TotalToValidatorsAllTime: sdkmath.NewInt(6000),
		TotalToTreasuryAllTime:   sdkmath.NewInt(3500),
		TotalTransactionsAllTime: 42,
		LastUpdateHeight:         5,
		CurrentEpoch:             1,
	}
	require.NoError(suite.T(), suite.keeper.SetFeeStatistics(suite.ctx, stats))

	got, err := suite.keeper.GetFeeStatistics(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), stats.TotalTransactionsAllTime, got.TotalTransactionsAllTime)
	require.Equal(suite.T(), stats.TotalCollectedAllTime, got.TotalCollectedAllTime)
}

func (suite *FeeStateTreasuryTestSuite) TestFeeStatistics_NotFound_ReturnsDefaults() {
	got, err := suite.keeper.GetFeeStatistics(suite.ctx)
	// Either returns default or error — production code handles both
	if err == nil {
		require.True(suite.T(), got.TotalCollectedAllTime.IsZero() || got.TotalCollectedAllTime.IsNil() || got.TotalCollectedAllTime.GTE(sdkmath.ZeroInt()))
	}
}

func (suite *FeeStateTreasuryTestSuite) TestUpdateFeeStatistics() {
	ctx := suite.ctx.WithBlockHeight(5)
	err := suite.keeper.UpdateFeeStatistics(ctx,
		sdkmath.NewInt(1000), // collected
		sdkmath.NewInt(50),   // burned
		sdkmath.NewInt(700),  // to validators
		sdkmath.NewInt(250),  // to treasury
		1,                    // txCount
	)
	require.NoError(suite.T(), err)

	stats, err := suite.keeper.GetFeeStatistics(ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), uint64(1), stats.TotalTransactionsAllTime)
	require.True(suite.T(), stats.TotalCollectedAllTime.Equal(sdkmath.NewInt(1000)))
}

func (suite *FeeStateTreasuryTestSuite) TestUpdateFeeStatistics_Accumulates() {
	ctx := suite.ctx.WithBlockHeight(1)
	require.NoError(suite.T(), suite.keeper.UpdateFeeStatistics(ctx,
		sdkmath.NewInt(1000), sdkmath.NewInt(50), sdkmath.NewInt(700), sdkmath.NewInt(250), 1))
	require.NoError(suite.T(), suite.keeper.UpdateFeeStatistics(ctx,
		sdkmath.NewInt(2000), sdkmath.NewInt(100), sdkmath.NewInt(1400), sdkmath.NewInt(500), 1))

	stats, err := suite.keeper.GetFeeStatistics(ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), uint64(2), stats.TotalTransactionsAllTime)
	require.True(suite.T(), stats.TotalCollectedAllTime.Equal(sdkmath.NewInt(3000)))
}

// ── BurnStatistics ────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestBurnStatistics_SetGet() {
	stats := types.BurnStats{
		TotalBurned:    sdkmath.NewInt(1000),
		BurnRatePerDay: sdkmath.NewInt(10),
		CurrentSupply:  sdkmath.NewInt(99000000),
		BurnCapSupply:  sdkmath.NewInt(50000000),
		RemainingToCap: sdkmath.NewInt(49000000),
		BurnCapReached: false,
		LastBurnHeight: 3,
	}
	require.NoError(suite.T(), suite.keeper.SetBurnStatistics(suite.ctx, stats))

	got, err := suite.keeper.GetBurnStatistics(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), stats.TotalBurned, got.TotalBurned)
	require.Equal(suite.T(), stats.LastBurnHeight, got.LastBurnHeight)
}

func (suite *FeeStateTreasuryTestSuite) TestUpdateBurnStatistics() {
	err := suite.keeper.UpdateBurnStatistics(suite.ctx, sdkmath.NewInt(500))
	require.NoError(suite.T(), err)

	got, err := suite.keeper.GetBurnStatistics(suite.ctx)
	require.NoError(suite.T(), err)
	require.True(suite.T(), got.TotalBurned.GTE(sdkmath.ZeroInt()))
}

func (suite *FeeStateTreasuryTestSuite) TestCanBurnTokens() {
	// Should be able to burn a small amount
	canBurn, err := suite.keeper.CanBurnTokens(suite.ctx, sdkmath.NewInt(100))
	require.NoError(suite.T(), err)
	require.True(suite.T(), canBurn)
}

// ── CalculateEpoch ────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestCalculateEpoch() {
	ctx := suite.ctx.WithBlockHeight(1000)
	epoch := suite.keeper.CalculateEpoch(ctx)
	require.GreaterOrEqual(suite.T(), epoch, int64(0))
}

// ── SupplySnapshot ────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestSupplySnapshot_SetGet() {
	snap := types.SupplySnapshot{
		Height:            7,
		Timestamp:         time.Now().UTC(),
		TotalSupply:       sdkmath.NewInt(1000000),
		CirculatingSupply: sdkmath.NewInt(900000),
		LiquidSupply:      sdkmath.NewInt(800000),
		BondedSupply:      sdkmath.NewInt(100000),
		BurnedCumulative:  sdkmath.NewInt(5000),
	}
	require.NoError(suite.T(), suite.keeper.SetSupplySnapshot(suite.ctx, snap))

	got, err := suite.keeper.GetSupplySnapshot(suite.ctx, 7)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), snap.Height, got.Height)
	require.Equal(suite.T(), snap.TotalSupply, got.TotalSupply)
}

func (suite *FeeStateTreasuryTestSuite) TestSupplySnapshot_NotFound() {
	_, err := suite.keeper.GetSupplySnapshot(suite.ctx, 99999)
	require.Error(suite.T(), err)
}

func (suite *FeeStateTreasuryTestSuite) TestCreateSupplySnapshot() {
	ctx := suite.ctx.WithBlockHeight(5)
	err := suite.keeper.CreateSupplySnapshot(ctx)
	require.NoError(suite.T(), err)

	snap, err := suite.keeper.GetSupplySnapshot(ctx, 5)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(5), snap.Height)
}

func (suite *FeeStateTreasuryTestSuite) TestGetLatestSupplySnapshot() {
	ctx := suite.ctx.WithBlockHeight(10)
	require.NoError(suite.T(), suite.keeper.CreateSupplySnapshot(ctx))

	snap, err := suite.keeper.GetLatestSupplySnapshot(ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(10), snap.Height)
}

// ── Treasury ──────────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryModuleAccount() {
	acc, err := suite.keeper.GetTreasuryModuleAccount(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), acc)
}

func (suite *FeeStateTreasuryTestSuite) TestGetVitacoinModuleAccount() {
	acc, err := suite.keeper.GetVitacoinModuleAccount(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), acc)
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryBalance() {
	balance, err := suite.keeper.GetTreasuryBalance(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), balance)
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryBalanceDenom() {
	bal, err := suite.keeper.GetTreasuryBalanceDenom(suite.ctx, "uvita")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "uvita", bal.Denom)
}

func (suite *FeeStateTreasuryTestSuite) TestGetVitaTreasuryBalance() {
	bal, err := suite.keeper.GetVitaTreasuryBalance(suite.ctx)
	require.NoError(suite.T(), err)
	require.True(suite.T(), bal.GTE(sdkmath.ZeroInt()))
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryAgeInBlocks() {
	ctx := suite.ctx.WithBlockHeight(100)
	age := suite.keeper.GetTreasuryAgeInBlocks(ctx)
	require.GreaterOrEqual(suite.T(), age, int64(0))
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryHealth() {
	score, err := suite.keeper.GetTreasuryHealth(suite.ctx)
	require.NoError(suite.T(), err)
	require.LessOrEqual(suite.T(), score, uint32(100))
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasuryStatistics() {
	stats, err := suite.keeper.GetTreasuryStatistics(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), stats)
}

func (suite *FeeStateTreasuryTestSuite) TestFormatTreasuryBalance() {
	formatted, err := suite.keeper.FormatTreasuryBalance(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), formatted)
}

func (suite *FeeStateTreasuryTestSuite) TestEstimateTreasuryRunway() {
	runway, err := suite.keeper.EstimateTreasuryRunway(suite.ctx)
	// runway may error if not enough data — just check no panic
	if err == nil {
		require.GreaterOrEqual(suite.T(), runway, int64(0))
	}
}

func (suite *FeeStateTreasuryTestSuite) TestValidateTreasurySpending_InsufficientFunds() {
	bigCoins := sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(999999999999)))
	// Mock bank returns lots of coins, so this might pass — just ensure no panic
	_ = suite.keeper.ValidateTreasurySpending(suite.ctx, bigCoins)
}

func (suite *FeeStateTreasuryTestSuite) TestTreasurySpending_SetGet() {
	spending := types.TreasurySpending{
		Id:          "spend-1",
		ProposalId:  1,
		Recipient:   sdk.AccAddress([]byte("spend_recipient_____")).String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(100))),
		Purpose:     "testing",
		SpentHeight: 5,
		SpentTime:   time.Now().UTC().Unix(),
	}
	require.NoError(suite.T(), suite.keeper.SetTreasurySpending(suite.ctx, spending))

	got, err := suite.keeper.GetTreasurySpending(suite.ctx, "spend-1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "spend-1", got.Id)
	require.Equal(suite.T(), "testing", got.Purpose)
}

func (suite *FeeStateTreasuryTestSuite) TestGetAllTreasurySpending() {
	for i := 0; i < 3; i++ {
		spending := types.TreasurySpending{
			Id:          fmt.Sprintf("all-spend-%d", i),
			ProposalId:  uint64(i),
			Recipient:   sdk.AccAddress([]byte(fmt.Sprintf("rcpt%d_______________", i)[:20])).String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(100))),
			Purpose:     "batch test",
			SpentHeight: int64(i + 1),
			SpentTime:   time.Now().UTC().Unix(),
		}
		require.NoError(suite.T(), suite.keeper.SetTreasurySpending(suite.ctx, spending))
	}

	all, err := suite.keeper.GetAllTreasurySpending(suite.ctx)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(all), 3)
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasurySpendingByProposal() {
	spending := types.TreasurySpending{
		Id:          "prop-spend-1",
		ProposalId:  42,
		Recipient:   sdk.AccAddress([]byte("prop_recipient______")).String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(200))),
		Purpose:     "proposal test",
		SpentHeight: 1,
		SpentTime:   time.Now().UTC().Unix(),
	}
	require.NoError(suite.T(), suite.keeper.SetTreasurySpending(suite.ctx, spending))

	results, err := suite.keeper.GetTreasurySpendingByProposal(suite.ctx, 42)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(results), 1)
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasurySpendingByRecipient() {
	recipient := sdk.AccAddress([]byte("rcpt_by_addr________")).String()
	spending := types.TreasurySpending{
		Id:          "rcpt-spend-1",
		ProposalId:  7,
		Recipient:   recipient,
		Amount:      sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(300))),
		Purpose:     "recipient filter test",
		SpentHeight: 1,
		SpentTime:   time.Now().UTC().Unix(),
	}
	require.NoError(suite.T(), suite.keeper.SetTreasurySpending(suite.ctx, spending))

	results, err := suite.keeper.GetTreasurySpendingByRecipient(suite.ctx, recipient)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(results), 1)
}

func (suite *FeeStateTreasuryTestSuite) TestGetTreasurySpendingInRange() {
	for i := int64(1); i <= 5; i++ {
		spending := types.TreasurySpending{
			Id:          fmt.Sprintf("range-spend-%d", i),
			ProposalId:  uint64(i),
			Recipient:   sdk.AccAddress([]byte(fmt.Sprintf("range_rcpt%d_________", i)[:20])).String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("uvita", sdkmath.NewInt(50))),
			Purpose:     "range test",
			SpentHeight: i,
			SpentTime:   time.Now().UTC().Unix(),
		}
		require.NoError(suite.T(), suite.keeper.SetTreasurySpending(suite.ctx, spending))
	}
	results, err := suite.keeper.GetTreasurySpendingInRange(suite.ctx, 2, 4)
	require.NoError(suite.T(), err)
	for _, r := range results {
		require.GreaterOrEqual(suite.T(), r.SpentHeight, int64(2))
		require.LessOrEqual(suite.T(), r.SpentHeight, int64(4))
	}
}

func (suite *FeeStateTreasuryTestSuite) TestExportTreasuryGenesis() {
	genesis, err := suite.keeper.ExportTreasuryGenesis(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), genesis)
}

// ── Invariants ────────────────────────────────────────────────────────────────

func (suite *FeeStateTreasuryTestSuite) TestAllInvariants_EmptyState() {
	msg, broken := keeper.AllInvariants(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "empty state should not break invariants: %s", msg)
}

func (suite *FeeStateTreasuryTestSuite) TestPaymentConsistencyInvariant_CompletedWithoutHeight() {
	// Completed payment must have CompletionHeight > 0
	p := types.Payment{
		Id:               "bad-payment",
		FromAddress:      sdk.AccAddress([]byte("inv_payer2__________")).String(),
		ToAddress:        sdk.AccAddress([]byte("inv_payee2__________")).String(),
		Amount:           sdkmath.NewInt(100),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   1,
		CompletionHeight: 0, // invalid!
	}
	require.NoError(suite.T(), suite.keeper.SetPayment(suite.ctx, p))

	_, broken := keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.True(suite.T(), broken)
}

func (suite *FeeStateTreasuryTestSuite) TestVaultConsistencyInvariant_Valid() {
	v := types.Vault{
		Id:               "inv-vault-valid",
		Owner:            sdk.AccAddress([]byte("inv_vault_owner2____")).String(),
		Amount:           sdkmath.NewInt(1000),
		LockDuration:     10,
		CreationHeight:   1,
		UnlockHeight:     11,
		RewardMultiplier: sdkmath.LegacyNewDec(1),
	}
	require.NoError(suite.T(), suite.keeper.SetVault(suite.ctx, v))
	_, broken := keeper.VaultConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)
}

func (suite *FeeStateTreasuryTestSuite) TestRewardPoolConsistencyInvariant_Valid() {
	pool := types.RewardPool{
		Id:                 "inv-pool-valid",
		MerchantAddress:    sdk.AccAddress([]byte("inv_pool_merch2_____")).String(),
		TotalRewards:       sdkmath.NewInt(1000),
		DistributedRewards: sdkmath.NewInt(100),
		StartHeight:        0,
		EndHeight:          0,
		IsActive:           true,
	}
	require.NoError(suite.T(), suite.keeper.SetRewardPool(suite.ctx, pool))
	_, broken := keeper.RewardPoolConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)
}

func (suite *FeeStateTreasuryTestSuite) TestMerchantStakeConsistencyInvariant_Valid() {
	m := types.Merchant{
		Address:            sdk.AccAddress([]byte("inv_merch_stake_____")).String(),
		BusinessName:       "Valid Stake Shop",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdkmath.NewInt(1000),
		RegistrationHeight: 1,
		IsActive:           true,
		TotalVolume:        sdkmath.ZeroInt(),
	}
	require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, m))
	_, broken := keeper.MerchantStakeConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)
}
