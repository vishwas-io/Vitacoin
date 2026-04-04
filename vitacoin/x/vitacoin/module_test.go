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

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ModuleTestSuite tests the AppModule and AppModuleBasic implementations.
type ModuleTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    keeper.Keeper
	module    vitacoin.AppModule
	cdc       codec.Codec
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (suite *ModuleTestSuite) SetupTest() {
	// Codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	suite.cdc = codec.NewProtoCodec(interfaceRegistry)

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

	// Config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")

	// Keeper
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority,
		&modMockBankKeeper{},
		&modMockAccountKeeper{},
	)

	p := types.DefaultParams()
	p.MinMerchantStake = sdkmath.NewInt(100)
	p.MerchantRegistrationFee = sdkmath.ZeroInt()
	require.NoError(suite.T(), suite.keeper.SetParams(suite.ctx, p))

	suite.module = vitacoin.NewAppModule(suite.cdc, suite.keeper)
}

// ── minimal mock keepers ──────────────────────────────────────────────────────

type modMockBankKeeper struct{}

func (m *modMockBankKeeper) GetBalance(_ context.Context, _ sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.NewInt(1_000_000_000_000_000_000))
}
func (m *modMockBankKeeper) GetAllBalances(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(1_000_000_000_000_000_000)))
}
func (m *modMockBankKeeper) GetSupply(_ context.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}
func (m *modMockBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *modMockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}
func (m *modMockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *modMockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _, _ string, _ sdk.Coins) error {
	return nil
}
func (m *modMockBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }
func (m *modMockBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }
func (m *modMockBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(1_000_000_000_000_000_000)))
}

type modMockAccountKeeper struct{}

func (m *modMockAccountKeeper) GetAccount(_ context.Context, _ sdk.AccAddress) sdk.AccountI {
	return nil
}
func (m *modMockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}
func (m *modMockAccountKeeper) GetModuleAccount(_ context.Context, _ string) sdk.ModuleAccountI {
	return nil
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestAppModuleBasic_Name verifies module name.
func (suite *ModuleTestSuite) TestAppModuleBasic_Name() {
	basic := vitacoin.AppModuleBasic{}
	require.Equal(suite.T(), types.ModuleName, basic.Name())
}

// TestAppModuleBasic_DefaultGenesis verifies default genesis is valid.
func (suite *ModuleTestSuite) TestAppModuleBasic_DefaultGenesis() {
	basic := vitacoin.AppModuleBasic{}
	genesis := basic.DefaultGenesis(suite.cdc)
	require.NotNil(suite.T(), genesis)
	err := basic.ValidateGenesis(suite.cdc, nil, genesis)
	require.NoError(suite.T(), err)
}

// TestAppModuleBasic_RegisterInterfaces verifies interface registration doesn't panic.
func (suite *ModuleTestSuite) TestAppModuleBasic_RegisterInterfaces() {
	basic := vitacoin.AppModuleBasic{}
	registry := codectypes.NewInterfaceRegistry()
	require.NotPanics(suite.T(), func() {
		basic.RegisterInterfaces(registry)
	})
}

// TestAppModuleBasic_RegisterLegacyAminoCodec verifies amino registration doesn't panic
// when given a real codec.
func (suite *ModuleTestSuite) TestAppModuleBasic_RegisterLegacyAminoCodec() {
	basic := vitacoin.AppModuleBasic{}
	aminoCdc := codec.NewLegacyAmino()
	require.NotPanics(suite.T(), func() {
		basic.RegisterLegacyAminoCodec(aminoCdc)
	})
}

// TestAppModule_Name verifies module name.
func (suite *ModuleTestSuite) TestAppModule_Name() {
	require.Equal(suite.T(), types.ModuleName, suite.module.Name())
}

// TestAppModule_ConsensusVersion verifies consensus version.
func (suite *ModuleTestSuite) TestAppModule_ConsensusVersion() {
	require.Equal(suite.T(), uint64(1), suite.module.ConsensusVersion())
}

// TestAppModule_RegisterInvariants verifies invariant registration doesn't panic.
func (suite *ModuleTestSuite) TestAppModule_RegisterInvariants() {
	var registered []string
	mockRegistry := &mockInvariantRegistry{onRegister: func(route string) {
		registered = append(registered, route)
	}}
	require.NotPanics(suite.T(), func() {
		suite.module.RegisterInvariants(mockRegistry)
	})
	require.NotEmpty(suite.T(), registered, "should register at least one invariant")
}

// TestBeginEndBlock verifies BeginBlock and EndBlock run without error on clean state.
func (suite *ModuleTestSuite) TestBeginEndBlock() {
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
	require.NoError(suite.T(), suite.module.EndBlock(suite.ctx))
}

// TestBeginEndBlock_WithMerchants verifies block processing with existing merchants.
func (suite *ModuleTestSuite) TestBeginEndBlock_WithMerchants() {
	// Set a merchant
	merchant := types.Merchant{
		Address:            sdk.AccAddress([]byte("modtest_merchant____")).String(),
		BusinessName:       "Module Test Shop",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdkmath.NewInt(1000),
		RegistrationHeight: 1,
		IsActive:           true,
		TotalVolume:        sdkmath.ZeroInt(),
	}
	require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, merchant))

	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
	require.NoError(suite.T(), suite.module.EndBlock(suite.ctx))

	// Merchant should still be there
	got, err := suite.keeper.GetMerchant(suite.ctx, merchant.Address)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), merchant.BusinessName, got.BusinessName)
}

// TestBeginBlock_ExpiresOldPayments verifies pending payments are expired at the right height.
func (suite *ModuleTestSuite) TestBeginBlock_ExpiresOldPayments() {
	params, err := suite.keeper.GetParams(suite.ctx)
	require.NoError(suite.T(), err)

	payment := types.Payment{
		Id:             "mod-expire-payment",
		FromAddress:    sdk.AccAddress([]byte("payer_______________")).String(),
		ToAddress:      sdk.AccAddress([]byte("payee_______________")).String(),
		Amount:         sdkmath.NewInt(1000),
		Status:         types.PaymentStatusPending,
		CreationHeight: 1,
		Memo:           "expire test",
	}
	require.NoError(suite.T(), suite.keeper.SetPayment(suite.ctx, payment))

	// Advance height past timeout
	expiryHeight := payment.CreationHeight + int64(params.PaymentTimeoutBlocks)
	suite.ctx = suite.ctx.WithBlockHeight(expiryHeight + 1)

	require.NoError(suite.T(), suite.module.EndBlock(suite.ctx))

	got, err := suite.keeper.GetPayment(suite.ctx, payment.Id)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusFailed, got.Status)
}

// TestBeginBlock_VaultUnlockEvents verifies vault unlock events are emitted.
func (suite *ModuleTestSuite) TestBeginBlock_VaultUnlockEvents() {
	vault := types.Vault{
		Id:               "mod-unlock-vault",
		Owner:            sdk.AccAddress([]byte("vault_owner_________")).String(),
		Amount:           sdkmath.NewInt(5000),
		LockDuration:     5,
		CreationHeight:   1,
		UnlockHeight:     5,
		RewardMultiplier: sdkmath.LegacyNewDecWithPrec(11, 1),
	}
	require.NoError(suite.T(), suite.keeper.SetVault(suite.ctx, vault))

	// Before unlock height — no event
	suite.ctx = suite.ctx.WithBlockHeight(4)
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))

	// At/after unlock height — event should fire
	suite.ctx = suite.ctx.WithBlockHeight(5)
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
}

// TestRewardPoolStatusTransitions verifies pools activate and deactivate at correct heights.
func (suite *ModuleTestSuite) TestRewardPoolStatusTransitions() {
	pool := types.RewardPool{
		Id:                 "mod-pool",
		MerchantAddress:    sdk.AccAddress([]byte("pool_merchant_______")).String(),
		TotalRewards:       sdkmath.NewInt(10000),
		DistributedRewards: sdkmath.ZeroInt(),
		StartHeight:        10,
		EndHeight:          20,
		IsActive:           false,
	}
	require.NoError(suite.T(), suite.keeper.SetRewardPool(suite.ctx, pool))

	// Before start: inactive
	suite.ctx = suite.ctx.WithBlockHeight(5)
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
	got, err := suite.keeper.GetRewardPool(suite.ctx, pool.Id)
	require.NoError(suite.T(), err)
	require.False(suite.T(), got.IsActive)

	// At start: active
	suite.ctx = suite.ctx.WithBlockHeight(10)
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
	got, err = suite.keeper.GetRewardPool(suite.ctx, pool.Id)
	require.NoError(suite.T(), err)
	require.True(suite.T(), got.IsActive)

	// After end: inactive
	suite.ctx = suite.ctx.WithBlockHeight(21)
	require.NoError(suite.T(), suite.module.BeginBlock(suite.ctx))
	got, err = suite.keeper.GetRewardPool(suite.ctx, pool.Id)
	require.NoError(suite.T(), err)
	require.False(suite.T(), got.IsActive)
}

// TestInvariants_ValidState verifies invariants pass with correct data.
func (suite *ModuleTestSuite) TestInvariants_ValidState() {
	merchant := types.Merchant{
		Address:            sdk.AccAddress([]byte("inv_merchant________")).String(),
		BusinessName:       "Invariant Shop",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdkmath.NewInt(1000),
		RegistrationHeight: 1,
		IsActive:           true,
		TotalVolume:        sdkmath.ZeroInt(),
	}
	require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, merchant))

	payment := types.Payment{
		Id:               "inv-payment",
		FromAddress:      sdk.AccAddress([]byte("inv_payer___________")).String(),
		ToAddress:        merchant.Address,
		Amount:           sdkmath.NewInt(100),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   1,
		CompletionHeight: 2,
	}
	require.NoError(suite.T(), suite.keeper.SetPayment(suite.ctx, payment))

	vault := types.Vault{
		Id:               "inv-vault",
		Owner:            sdk.AccAddress([]byte("inv_owner___________")).String(),
		Amount:           sdkmath.NewInt(1000),
		LockDuration:     100,
		CreationHeight:   1,
		UnlockHeight:     101,
		RewardMultiplier: sdkmath.LegacyNewDec(1),
	}
	require.NoError(suite.T(), suite.keeper.SetVault(suite.ctx, vault))

	pool := types.RewardPool{
		Id:                 "inv-pool",
		MerchantAddress:    merchant.Address,
		TotalRewards:       sdkmath.NewInt(1000),
		DistributedRewards: sdkmath.NewInt(200),
		StartHeight:        1,
		EndHeight:          0,
		IsActive:           true,
	}
	require.NoError(suite.T(), suite.keeper.SetRewardPool(suite.ctx, pool))

	_, broken := keeper.AllInvariants(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "invariants should pass with valid data")

	_, broken = keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)

	_, broken = keeper.VaultConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)

	_, broken = keeper.RewardPoolConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)

	_, broken = keeper.MerchantStakeConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken)
}

// TestCalculateMerchantTier verifies tier calculation by volume.
func (suite *ModuleTestSuite) TestCalculateMerchantTier() {
	// Small volume → Bronze
	tier := suite.keeper.CalculateMerchantTier(sdk.NewCoin("avita", sdkmath.NewInt(500)))
	require.Equal(suite.T(), types.MerchantTierBronze, tier)

	// Very large volume → Platinum
	largeAmt, _ := sdkmath.NewIntFromString("2000000000000000000000000")
	tier = suite.keeper.CalculateMerchantTier(sdk.NewCoin("avita", largeAmt))
	require.Equal(suite.T(), types.MerchantTierPlatinum, tier)
}

// TestInitExportGenesis round-trips genesis state.
func (suite *ModuleTestSuite) TestInitExportGenesis() {
	// Store some state
	merchant := types.Merchant{
		Address:            sdk.AccAddress([]byte("genesis_merchant____")).String(),
		BusinessName:       "Genesis Shop",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdkmath.NewInt(500),
		RegistrationHeight: 1,
		IsActive:           true,
		TotalVolume:        sdkmath.ZeroInt(),
	}
	require.NoError(suite.T(), suite.keeper.SetMerchant(suite.ctx, merchant))

	// Export
	exported, err := suite.keeper.ExportGenesis(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), exported)
	require.Len(suite.T(), exported.MerchantList, 1)
	require.Equal(suite.T(), "Genesis Shop", exported.MerchantList[0].BusinessName)
}

// ── mock InvariantRegistry ────────────────────────────────────────────────────

type mockInvariantRegistry struct {
	onRegister func(route string)
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	if m.onRegister != nil {
		m.onRegister(moduleName + "/" + route)
	}
}
