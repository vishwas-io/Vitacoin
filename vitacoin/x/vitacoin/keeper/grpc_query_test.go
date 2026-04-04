package keeper_test

// grpc_query_test.go — tests for all gRPC query handlers to boost coverage.

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

type GRPCQueryTestSuite struct {
	KeeperTestSuite // reuse the setup from keeper_test.go
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) addr(seed string) string {
	raw := []byte(seed)
	if len(raw) < 20 {
		padded := make([]byte, 20)
		copy(padded, raw)
		raw = padded
	}
	return sdk.AccAddress(raw[:20]).String()
}

// ── Params ───────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryParams_NilRequest() {
	_, err := suite.queryServer.Params(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryParams_OK() {
	resp, err := suite.queryServer.Params(suite.ctx, &types.QueryParamsRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── Merchant ──────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryMerchant_NilRequest() {
	_, err := suite.queryServer.Merchant(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchant_EmptyAddress() {
	_, err := suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{Address: ""})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchant_InvalidAddress() {
	_, err := suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{Address: "not-bech32"})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchant_NotFound() {
	addr := suite.addr("qry_merchant_404____")
	_, err := suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{Address: addr})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchant_OK() {
	addr := suite.addr("qry_merchant_ok_____")
	m := types.Merchant{
		Address:            addr,
		BusinessName:       "Query Shop",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdkmath.NewInt(1000),
		RegistrationHeight: 1,
		IsActive:           true,
		TotalVolume:        sdkmath.ZeroInt(),
	}
	require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, m))

	resp, err := suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{Address: addr})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Query Shop", resp.Merchant.BusinessName)
}

// ── MerchantAll ───────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryMerchantAll_NilRequest() {
	_, err := suite.queryServer.MerchantAll(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchantAll_Empty() {
	resp, err := suite.queryServer.MerchantAll(suite.ctx, &types.QueryAllMerchantRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

func (suite *GRPCQueryTestSuite) TestQueryMerchantAll_Multiple() {
	for i := byte(0); i < 3; i++ {
		seed := make([]byte, 20)
		seed[0] = 'A' + i
		addr := sdk.AccAddress(seed).String()
		m := types.Merchant{
			Address:            addr,
			BusinessName:       "Shop",
			Tier:               types.MerchantTierBronze,
			StakeAmount:        sdkmath.NewInt(1000),
			RegistrationHeight: 1,
			IsActive:           true,
			TotalVolume:        sdkmath.ZeroInt(),
		}
		require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, m))
	}
	resp, err := suite.queryServer.MerchantAll(suite.ctx, &types.QueryAllMerchantRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(resp.Merchants), 3)
}

// ── Payment ───────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryPayment_NilRequest() {
	_, err := suite.queryServer.Payment(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryPayment_EmptyID() {
	_, err := suite.queryServer.Payment(suite.ctx, &types.QueryPaymentRequest{Id: ""})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryPayment_NotFound() {
	_, err := suite.queryServer.Payment(suite.ctx, &types.QueryPaymentRequest{Id: "nonexistent-pay"})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryPayment_OK() {
	p := types.Payment{
		Id:             "qry-pay-ok",
		FromAddress:    suite.addr("qry_payer___________"),
		ToAddress:      suite.addr("qry_payee___________"),
		Amount:         sdkmath.NewInt(500),
		Status:         types.PaymentStatusPending,
		CreationHeight: 1,
	}
	require.NoError(suite.T(), suite.keeper.SetPayment(suite.ctx, p))

	resp, err := suite.queryServer.Payment(suite.ctx, &types.QueryPaymentRequest{Id: "qry-pay-ok"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "qry-pay-ok", resp.Payment.Id)
}

// ── PaymentAll ────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryPaymentAll_NilRequest() {
	_, err := suite.queryServer.PaymentAll(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryPaymentAll_OK() {
	resp, err := suite.queryServer.PaymentAll(suite.ctx, &types.QueryAllPaymentRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── Vault ─────────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryVault_NilRequest() {
	_, err := suite.queryServer.Vault(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryVault_EmptyID() {
	_, err := suite.queryServer.Vault(suite.ctx, &types.QueryVaultRequest{Id: ""})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryVault_NotFound() {
	_, err := suite.queryServer.Vault(suite.ctx, &types.QueryVaultRequest{Id: "no-vault"})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryVault_OK() {
	v := types.Vault{
		Id:               "qry-vault-ok",
		Owner:            suite.addr("qry_vault_owner_____"),
		Amount:           sdkmath.NewInt(2000),
		LockDuration:     100,
		CreationHeight:   1,
		UnlockHeight:     101,
		RewardMultiplier: sdkmath.LegacyNewDec(1),
	}
	require.NoError(suite.T(), suite.keeper.SetVault(suite.ctx, v))

	resp, err := suite.queryServer.Vault(suite.ctx, &types.QueryVaultRequest{Id: "qry-vault-ok"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "qry-vault-ok", resp.Vault.Id)
}

// ── VaultAll ──────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryVaultAll_NilRequest() {
	_, err := suite.queryServer.VaultAll(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryVaultAll_OK() {
	resp, err := suite.queryServer.VaultAll(suite.ctx, &types.QueryAllVaultRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── RewardPool ────────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryRewardPool_NilRequest() {
	_, err := suite.queryServer.RewardPool(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryRewardPool_EmptyID() {
	_, err := suite.queryServer.RewardPool(suite.ctx, &types.QueryRewardPoolRequest{Id: ""})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryRewardPool_NotFound() {
	_, err := suite.queryServer.RewardPool(suite.ctx, &types.QueryRewardPoolRequest{Id: "no-pool"})
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryRewardPool_OK() {
	pool := types.RewardPool{
		Id:                 "qry-pool-ok",
		MerchantAddress:    suite.addr("qry_pool_merchant___"),
		TotalRewards:       sdkmath.NewInt(3000),
		DistributedRewards: sdkmath.ZeroInt(),
		StartHeight:        1,
		EndHeight:          0,
		IsActive:           true,
	}
	require.NoError(suite.T(), suite.keeper.SetRewardPool(suite.ctx, pool))

	resp, err := suite.queryServer.RewardPool(suite.ctx, &types.QueryRewardPoolRequest{Id: "qry-pool-ok"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "qry-pool-ok", resp.Pool.Id)
}

// ── RewardPoolAll ─────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryRewardPoolAll_NilRequest() {
	_, err := suite.queryServer.RewardPoolAll(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryRewardPoolAll_OK() {
	resp, err := suite.queryServer.RewardPoolAll(suite.ctx, &types.QueryAllRewardPoolRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── FeeStatistics ─────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryFeeStatistics_NilRequest() {
	_, err := suite.queryServer.FeeStatistics(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryFeeStatistics_OK() {
	resp, err := suite.queryServer.FeeStatistics(suite.ctx, &types.QueryFeeStatisticsRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── BurnStatistics ────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryBurnStatistics_NilRequest() {
	_, err := suite.queryServer.BurnStatistics(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryBurnStatistics_OK() {
	resp, err := suite.queryServer.BurnStatistics(suite.ctx, &types.QueryBurnStatisticsRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── FeeAccumulator ────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryFeeAccumulator_NilRequest() {
	_, err := suite.queryServer.FeeAccumulator(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryFeeAccumulator_OK() {
	resp, err := suite.queryServer.FeeAccumulator(suite.ctx, &types.QueryFeeAccumulatorRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── SupplySnapshotLatest ──────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQuerySupplySnapshotLatest_NilRequest() {
	_, err := suite.queryServer.SupplySnapshotLatest(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQuerySupplySnapshotLatest_OK() {
	// Use a context with height>0 so GetLatestSupplySnapshot loop can run
	ctx := suite.ctx.WithBlockHeight(1)
	resp, err := suite.queryServer.SupplySnapshotLatest(ctx, &types.QuerySupplySnapshotLatestRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── SupplySnapshot ────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQuerySupplySnapshot_NilRequest() {
	_, err := suite.queryServer.SupplySnapshot(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQuerySupplySnapshot_NotFound() {
	// Height with no snapshot → expect error
	_, err := suite.queryServer.SupplySnapshot(suite.ctx, &types.QuerySupplySnapshotRequest{Height: 9999999})
	require.Error(suite.T(), err)
}

// ── TreasuryBalance ───────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryTreasuryBalance_NilRequest() {
	_, err := suite.keeper.TreasuryBalance(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryTreasuryBalance_OK() {
	resp, err := suite.keeper.TreasuryBalance(suite.ctx, &types.QueryTreasuryBalanceRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}

// ── TreasuryHealth ────────────────────────────────────────────────────────────

func (suite *GRPCQueryTestSuite) TestQueryTreasuryHealth_NilRequest() {
	_, err := suite.keeper.TreasuryHealth(suite.ctx, nil)
	require.Error(suite.T(), err)
}

func (suite *GRPCQueryTestSuite) TestQueryTreasuryHealth_OK() {
	resp, err := suite.keeper.TreasuryHealth(suite.ctx, &types.QueryTreasuryHealthRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
}
