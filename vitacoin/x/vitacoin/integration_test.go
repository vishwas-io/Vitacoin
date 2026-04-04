package vitacoin_test

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// IntegrationTestSuite tests cross-cutting module behaviour end-to-end.
type IntegrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	keeper      keeper.Keeper
	msgServer   types.MsgServer
	queryServer types.QueryServer
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Configure bech32 prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")
	config.SetBech32PrefixForValidator("vitavaloper", "vitavaloperpub")
	config.SetBech32PrefixForConsensusNode("vitavalcons", "vitavalconspub")

	// Codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	// Context
	suite.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	// Keeper
	suite.keeper = keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		&integMockBankKeeper{},
		&integMockAccountKeeper{},
	)

	// Low-threshold params so tests can use small amounts
	p := types.DefaultParams()
	p.MinMerchantStake = sdkmath.NewInt(1000)
	p.MerchantRegistrationFee = sdkmath.ZeroInt()
	p.MinProtocolFee = sdkmath.ZeroInt()
	p.MaxProtocolFee = sdkmath.NewInt(1_000_000_000_000_000_000)
	require.NoError(suite.T(), suite.keeper.SetParams(suite.ctx, p))

	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)
}

// ── minimal mock keepers (use context.Context as the interface requires) ──────

type integMockBankKeeper struct{}

func (m *integMockBankKeeper) GetBalance(_ context.Context, _ sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.NewInt(1_000_000_000_000_000_000))
}
func (m *integMockBankKeeper) GetAllBalances(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(1_000_000_000_000_000_000)))
}
func (m *integMockBankKeeper) GetSupply(_ context.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}
func (m *integMockBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *integMockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}
func (m *integMockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *integMockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _, _ string, _ sdk.Coins) error {
	return nil
}
func (m *integMockBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }
func (m *integMockBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }
func (m *integMockBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(1_000_000_000_000_000_000)))
}

type integMockAccountKeeper struct{}

func (m *integMockAccountKeeper) GetAccount(_ context.Context, _ sdk.AccAddress) sdk.AccountI {
	return nil
}
func (m *integMockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}
func (m *integMockAccountKeeper) GetModuleAccount(_ context.Context, _ string) sdk.ModuleAccountI {
	return nil
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestMerchantLifecycle verifies register → update → query flow.
func (suite *IntegrationTestSuite) TestMerchantLifecycle() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	addr := sdk.AccAddress([]byte("merchant1___________"))

	// Register
	_, err := suite.msgServer.RegisterMerchant(ctx, &types.MsgRegisterMerchant{
		Sender:       addr.String(),
		BusinessName: "Coffee Shop",
		StakeAmount:  sdkmath.NewInt(5000),
	})
	require.NoError(suite.T(), err)

	// Query
	merchant, err := suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Coffee Shop", merchant.BusinessName)
	require.True(suite.T(), merchant.IsActive)

	// Update
	_, err = suite.msgServer.UpdateMerchant(ctx, &types.MsgUpdateMerchant{
		Sender:          addr.String(),
		BusinessName:    "Coffee Shop - Downtown",
		AdditionalStake: sdkmath.NewInt(1000),
	})
	require.NoError(suite.T(), err)

	merchant, err = suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Coffee Shop - Downtown", merchant.BusinessName)
}

// TestPaymentFlow verifies create → complete → refund transitions.
func (suite *IntegrationTestSuite) TestPaymentFlow() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	merchantAddr := sdk.AccAddress([]byte("merchant_pay________"))
	customerAddr := sdk.AccAddress([]byte("customer_pay________"))

	// Register merchant
	_, err := suite.msgServer.RegisterMerchant(ctx, &types.MsgRegisterMerchant{
		Sender:       merchantAddr.String(),
		BusinessName: "E-commerce Store",
		StakeAmount:  sdkmath.NewInt(5000),
	})
	require.NoError(suite.T(), err)

	// Create payment
	createResp, err := suite.msgServer.CreatePayment(ctx, &types.MsgCreatePayment{
		Sender:          customerAddr.String(),
		MerchantAddress: merchantAddr.String(),
		Amount:          sdkmath.NewInt(1000),
		Memo:            "integration test",
	})
	require.NoError(suite.T(), err)
	paymentID := createResp.PaymentId
	require.NotEmpty(suite.T(), paymentID)

	// Verify pending
	payment, err := suite.keeper.GetPayment(suite.ctx, paymentID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusPending, payment.Status)

	// Complete
	_, err = suite.msgServer.CompletePayment(ctx, &types.MsgCompletePayment{
		Sender:    merchantAddr.String(),
		PaymentId: paymentID,
	})
	require.NoError(suite.T(), err)

	payment, err = suite.keeper.GetPayment(suite.ctx, paymentID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusCompleted, payment.Status)

	// Refund
	_, err = suite.msgServer.RefundPayment(ctx, &types.MsgRefundPayment{
		Sender:    merchantAddr.String(),
		PaymentId: paymentID,
		Reason:    "customer request",
	})
	require.NoError(suite.T(), err)

	payment, err = suite.keeper.GetPayment(suite.ctx, paymentID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusRefunded, payment.Status)
}

// TestVaultOperations verifies vault create and withdraw flow.
func (suite *IntegrationTestSuite) TestVaultOperations() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	userAddr := sdk.AccAddress([]byte("vault_user__________"))

	// Create vault
	createResp, err := suite.msgServer.CreateVault(ctx, &types.MsgCreateVault{
		Sender:       userAddr.String(),
		Amount:       sdkmath.NewInt(100_000),
		LockDuration: 100,
	})
	require.NoError(suite.T(), err)
	vaultID := createResp.VaultId
	require.NotEmpty(suite.T(), vaultID)

	// Verify vault state
	vault, err := suite.keeper.GetVault(suite.ctx, vaultID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), userAddr.String(), vault.Owner)

	// Advance past lock height and withdraw
	suite.ctx = suite.ctx.WithBlockHeight(int64(vault.UnlockHeight) + 1)
	ctx = sdk.UnwrapSDKContext(suite.ctx)

	_, err = suite.msgServer.WithdrawVault(ctx, &types.MsgWithdrawVault{
		Sender:  userAddr.String(),
		VaultId: vaultID,
	})
	require.NoError(suite.T(), err)
}

// TestRewardDistribution verifies reward pool create and distribute flow.
func (suite *IntegrationTestSuite) TestRewardDistribution() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	recipient1 := sdk.AccAddress([]byte("reward_recv1________"))
	recipient2 := sdk.AccAddress([]byte("reward_recv2________"))

	// Register recipients as merchants (needed for eligibility)
	for _, addr := range []sdk.AccAddress{recipient1, recipient2} {
		_, err := suite.msgServer.RegisterMerchant(ctx, &types.MsgRegisterMerchant{
			Sender:       addr.String(),
			BusinessName: "Reward Merchant",
			StakeAmount:  sdkmath.NewInt(1000),
		})
		require.NoError(suite.T(), err)
	}

	// Create reward pool — sender must be a registered merchant
	poolResp, err := suite.msgServer.CreateRewardPool(ctx, &types.MsgCreateRewardPool{
		Sender:         recipient1.String(),
		TotalRewards:   sdkmath.NewInt(5000),
		DurationBlocks: 86400,
	})
	require.NoError(suite.T(), err)
	poolID := poolResp.PoolId
	require.NotEmpty(suite.T(), poolID)

	// Activate pool so distribution can proceed
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	ctx = sdk.UnwrapSDKContext(suite.ctx)

	// Distribute rewards — sender must be the pool's merchant address
	_, err = suite.msgServer.DistributeRewards(ctx, &types.MsgDistributeRewards{
		Sender:     recipient1.String(),
		PoolId:     poolID,
		Recipients: []string{recipient1.String(), recipient2.String()},
		Amounts:    []sdkmath.Int{sdkmath.NewInt(3000), sdkmath.NewInt(2000)},
	})
	require.NoError(suite.T(), err)

	// Pool should be depleted
	pool, err := suite.keeper.GetRewardPool(suite.ctx, poolID)
	require.NoError(suite.T(), err)
	require.True(suite.T(), pool.TotalRewards.Sub(pool.DistributedRewards).IsZero())
}

// TestGovernanceParamUpdate verifies that governance can update params.
func (suite *IntegrationTestSuite) TestGovernanceParamUpdate() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	newParams := types.DefaultParams()
	newParams.MinMerchantStake = sdkmath.NewInt(9999)

	_, err := suite.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: govAddr,
		Params:    newParams,
	})
	require.NoError(suite.T(), err)

	updated, err := suite.keeper.GetParams(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), sdkmath.NewInt(9999), updated.MinMerchantStake)
}

// TestQueryEndpoints exercises all major query paths with real state.
func (suite *IntegrationTestSuite) TestQueryEndpoints() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	merchantAddr := sdk.AccAddress([]byte("query_merchant______"))
	customerAddr := sdk.AccAddress([]byte("query_customer______"))

	// Register merchant
	_, err := suite.msgServer.RegisterMerchant(ctx, &types.MsgRegisterMerchant{
		Sender:       merchantAddr.String(),
		BusinessName: "Query Test Merchant",
		StakeAmount:  sdkmath.NewInt(5000),
	})
	require.NoError(suite.T(), err)

	// Create payment
	createResp, err := suite.msgServer.CreatePayment(ctx, &types.MsgCreatePayment{
		Sender:          customerAddr.String(),
		MerchantAddress: merchantAddr.String(),
		Amount:          sdkmath.NewInt(1000),
		Memo:            "query test",
	})
	require.NoError(suite.T(), err)

	// Query Params
	paramsResp, err := suite.queryServer.Params(suite.ctx, &types.QueryParamsRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), paramsResp.Params)

	// Query Merchant
	merchantResp, err := suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{
		Address: merchantAddr.String(),
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Query Test Merchant", merchantResp.Merchant.BusinessName)

	// Query AllMerchants
	allMerchantsResp, err := suite.queryServer.MerchantAll(suite.ctx, &types.QueryAllMerchantRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(allMerchantsResp.Merchants), 1)

	// Query Payment
	paymentResp, err := suite.queryServer.Payment(suite.ctx, &types.QueryPaymentRequest{
		Id: createResp.PaymentId,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), createResp.PaymentId, paymentResp.Payment.Id)

	// Query AllPayments
	allPaymentsResp, err := suite.queryServer.PaymentAll(suite.ctx, &types.QueryAllPaymentRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(allPaymentsResp.Payments), 1)

	// Non-existent queries should error
	_, err = suite.queryServer.Merchant(suite.ctx, &types.QueryMerchantRequest{Address: "nonexistent"})
	require.Error(suite.T(), err)

	_, err = suite.queryServer.Payment(suite.ctx, &types.QueryPaymentRequest{Id: "nonexistent"})
	require.Error(suite.T(), err)
}

// TestBlockLifecycle verifies BeginBlocker and EndBlocker run without error.
func (suite *IntegrationTestSuite) TestBlockLifecycle() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Register a merchant so there is actual state
	addr := sdk.AccAddress([]byte("lifecycle_merchant__"))
	_, err := suite.msgServer.RegisterMerchant(ctx, &types.MsgRegisterMerchant{
		Sender:       addr.String(),
		BusinessName: "Lifecycle Merchant",
		StakeAmount:  sdkmath.NewInt(5000),
	})
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), suite.keeper.BeginBlocker(suite.ctx))
	require.NoError(suite.T(), suite.keeper.EndBlocker(suite.ctx))

	// State must still be consistent
	merchant, err := suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.NoError(suite.T(), err)
	require.True(suite.T(), merchant.IsActive)
}
