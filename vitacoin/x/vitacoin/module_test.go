package vitacoin_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

type ModuleTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    keeper.Keeper
	module    vitacoin.AppModule
	cdc       codec.Codec
	storeKey  *storetypes.KVStoreKey
	authority string
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (suite *ModuleTestSuite) SetupTest() {
	suite.storeKey = sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), suite.storeKey, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})
	
	encCfg := moduletestutil.MakeTestEncodingConfig(vitacoin.AppModuleBasic{})
	suite.cdc = encCfg.Codec
	
	// Set up authority address for testing
	suite.authority = "cosmos1test_authority_address_12345678901234567890"

	// Create store service from runtime
	storeService := runtime.NewKVStoreService(suite.storeKey)
	
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeService,
		log.NewNopLogger(),
		suite.authority,
	)

	suite.module = vitacoin.NewAppModule(suite.cdc, suite.keeper)
}

func (suite *ModuleTestSuite) TestAppModuleBasic() {
	basic := vitacoin.AppModuleBasic{}

	// Test module name
	require.Equal(suite.T(), types.ModuleName, basic.Name())

	// Test interfaces registration
	encCfg := moduletestutil.MakeTestEncodingConfig(vitacoin.AppModuleBasic{})
	registry := encCfg.InterfaceRegistry
	require.NotPanics(suite.T(), func() {
		basic.RegisterInterfaces(registry)
	})

	// Test default genesis
	genesis := basic.DefaultGenesis(suite.cdc)
	require.NotNil(suite.T(), genesis)

	// Test genesis validation
	err := basic.ValidateGenesis(suite.cdc, nil, genesis)
	require.NoError(suite.T(), err)

	// Test legacy amino codec registration
	require.NotPanics(suite.T(), func() {
		basic.RegisterLegacyAminoCodec(nil)
	})
}

func (suite *ModuleTestSuite) TestAppModule() {
	// Test module name
	require.Equal(suite.T(), types.ModuleName, suite.module.Name())

	// Test consensus version
	require.Equal(suite.T(), uint64(1), suite.module.ConsensusVersion())

	// Test invariants registration - should not panic
	ir := sdk.NewInvariantRegistry()
	require.NotPanics(suite.T(), func() {
		suite.module.RegisterInvariants(ir)
	})

	// Test query routes
	routes := suite.module.LegacyQuerierHandler(nil)
	require.Nil(suite.T(), routes) // Should be nil as we use gRPC

	// Test services registration
	cfg := &moduletestutil.TestConfigurator{}
	require.NotPanics(suite.T(), func() {
		suite.module.RegisterServices(cfg)
	})
}

func (suite *ModuleTestSuite) TestBeginEndBlock() {
	// Setup initial state with a payment that should expire
	payment := types.Payment{
		Id:             "test-payment-1",
		FromAddress:    "cosmos1test1",
		ToAddress:      "cosmos1test2", 
		Amount:         sdk.NewCoin("uvita", sdk.NewInt(1000)),
		Fee:            sdk.NewCoin("uvita", sdk.NewInt(10)),
		Status:         types.PaymentStatusPending,
		CreationHeight: 1,
		ExpiryHeight:   5, // Will expire at height 5
	}

	err := suite.keeper.CreatePayment(suite.ctx, payment)
	require.NoError(suite.T(), err)

	// Test BeginBlock at height 1 - payment should still be pending
	suite.ctx = suite.ctx.WithBlockHeight(1)
	require.NotPanics(suite.T(), func() {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	})

	retrievedPayment, err := suite.keeper.GetPayment(suite.ctx, "test-payment-1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusPending, retrievedPayment.Status)

	// Test BeginBlock at height 6 - payment should be expired
	suite.ctx = suite.ctx.WithBlockHeight(6)
	require.NotPanics(suite.T(), func() {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	})

	retrievedPayment, err = suite.keeper.GetPayment(suite.ctx, "test-payment-1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusExpired, retrievedPayment.Status)

	// Test EndBlock - should not panic
	require.NotPanics(suite.T(), func() {
		suite.module.EndBlock(suite.ctx, abci.RequestEndBlock{})
	})
}

func (suite *ModuleTestSuite) TestInvariants() {
	// Create test data
	merchant := types.Merchant{
		Address:            "cosmos1merchant",
		BusinessName:       "Test Business",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdk.NewCoin("uvita", sdk.NewInt(1000)),
		TotalVolume:        sdk.NewCoin("uvita", sdk.NewInt(500)), // Low volume = Bronze tier
		RegistrationHeight: 1,
	}

	err := suite.keeper.CreateMerchant(suite.ctx, merchant)
	require.NoError(suite.T(), err)

	payment := types.Payment{
		Id:               "test-payment",
		FromAddress:      "cosmos1test1",
		ToAddress:        "cosmos1merchant",
		Amount:           sdk.NewCoin("uvita", sdk.NewInt(100)),
		Fee:              sdk.NewCoin("uvita", sdk.NewInt(1)),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   1,
		CompletionHeight: 2,
	}

	err = suite.keeper.CreatePayment(suite.ctx, payment)
	require.NoError(suite.T(), err)

	vault := types.Vault{
		Id:               "test-vault",
		Owner:            "cosmos1test1",
		Amount:           sdk.NewCoin("uvita", sdk.NewInt(1000)),
		CreationHeight:   1,
		UnlockHeight:     100,
		RewardMultiplier: sdk.NewDecWithPrec(15, 1), // 1.5x
		Withdrawn:        false,
	}

	err = suite.keeper.CreateVault(suite.ctx, vault)
	require.NoError(suite.T(), err)

	rewardPool := types.RewardPool{
		Id:                 "test-pool",
		MerchantAddress:    "cosmos1merchant",
		TotalRewards:       sdk.NewCoin("uvita", sdk.NewInt(1000)),
		DistributedRewards: sdk.NewCoin("uvita", sdk.NewInt(200)),
		StartHeight:        1,
		EndHeight:          0,
		Active:             true,
	}

	err = suite.keeper.CreateRewardPool(suite.ctx, rewardPool)
	require.NoError(suite.T(), err)

	// Test all invariants
	invariants := keeper.AllInvariants(suite.keeper)
	_, broken := invariants(suite.ctx)
	require.False(suite.T(), broken, "Invariants should not be broken with valid data")

	// Test individual invariants
	_, broken = keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Payment invariant should not be broken")

	_, broken = keeper.VaultConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Vault invariant should not be broken")

	_, broken = keeper.RewardPoolConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Reward pool invariant should not be broken")

	_, broken = keeper.MerchantStakeConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Merchant stake invariant should not be broken")
}

func (suite *ModuleTestSuite) TestBrokenInvariants() {
	// Create invalid payment data to break invariants
	invalidPayment := types.Payment{
		Id:               "invalid-payment",
		FromAddress:      "invalid-address", // Invalid bech32 address
		ToAddress:        "cosmos1merchant",
		Amount:           sdk.Coin{}, // Invalid amount
		Fee:              sdk.NewCoin("uvita", sdk.NewInt(-1)), // Negative fee
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   5,
		CompletionHeight: 2, // Completion before creation
	}

	// Manually store invalid data to test invariant detection
	store := suite.ctx.KVStore(suite.storeKey)
	bz, err := suite.cdc.Marshal(&invalidPayment)
	require.NoError(suite.T(), err)
	store.Set(types.PaymentKey(invalidPayment.Id), bz)

	// Test payment invariant with invalid data
	_, broken := keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.True(suite.T(), broken, "Payment invariant should be broken with invalid data")

	// Test that all invariants detect the broken state
	_, broken = keeper.AllInvariants(suite.keeper)(suite.ctx)
	require.True(suite.T(), broken, "All invariants should detect broken state")
}

func (suite *ModuleTestSuite) TestMerchantTierCalculation() {
	// Test merchant tier calculation with different volumes
	lowVolume := sdk.NewCoin("uvita", sdk.NewInt(500))
	tier := suite.keeper.CalculateMerchantTier(lowVolume)
	require.Equal(suite.T(), types.MerchantTierBronze, tier)

	mediumVolume := sdk.NewCoin("uvita", sdk.NewInt(5000))
	tier = suite.keeper.CalculateMerchantTier(mediumVolume)
	require.Equal(suite.T(), types.MerchantTierSilver, tier)

	highVolume := sdk.NewCoin("uvita", sdk.NewInt(50000))
	tier = suite.keeper.CalculateMerchantTier(highVolume)
	require.Equal(suite.T(), types.MerchantTierGold, tier)
}

func (suite *ModuleTestSuite) TestVaultUnlocking() {
	// Create a vault that should unlock
	vault := types.Vault{
		Id:               "unlock-vault",
		Owner:            "cosmos1test1",
		Amount:           sdk.NewCoin("uvita", sdk.NewInt(1000)),
		CreationHeight:   1,
		UnlockHeight:     5,
		RewardMultiplier: sdk.NewDecWithPrec(15, 1),
		Withdrawn:        false,
	}

	err := suite.keeper.CreateVault(suite.ctx, vault)
	require.NoError(suite.T(), err)

	// At height 4, vault should not be unlocked
	suite.ctx = suite.ctx.WithBlockHeight(4)
	suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})

	retrievedVault, err := suite.keeper.GetVault(suite.ctx, "unlock-vault")
	require.NoError(suite.T(), err)
	require.False(suite.T(), retrievedVault.Withdrawn)

	// At height 5, vault should be available for withdrawal
	suite.ctx = suite.ctx.WithBlockHeight(5)
	suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})

	// The vault should still exist but be unlocked (not automatically withdrawn)
	retrievedVault, err = suite.keeper.GetVault(suite.ctx, "unlock-vault")
	require.NoError(suite.T(), err)
	require.False(suite.T(), retrievedVault.Withdrawn) // Only unlocked, not withdrawn
}

func (suite *ModuleTestSuite) TestRewardPoolActivation() {
	// Create a reward pool with future activation
	rewardPool := types.RewardPool{
		Id:                 "future-pool",
		MerchantAddress:    "cosmos1merchant",
		TotalRewards:       sdk.NewCoin("uvita", sdk.NewInt(1000)),
		DistributedRewards: sdk.NewCoin("uvita", sdk.NewInt(0)),
		StartHeight:        10,
		EndHeight:          20,
		Active:             false,
	}

	err := suite.keeper.CreateRewardPool(suite.ctx, rewardPool)
	require.NoError(suite.T(), err)

	// At height 5, pool should not be active
	suite.ctx = suite.ctx.WithBlockHeight(5)
	suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})

	retrievedPool, err := suite.keeper.GetRewardPool(suite.ctx, "future-pool")
	require.NoError(suite.T(), err)
	require.False(suite.T(), retrievedPool.Active)

	// At height 10, pool should be active
	suite.ctx = suite.ctx.WithBlockHeight(10)
	suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})

	retrievedPool, err = suite.keeper.GetRewardPool(suite.ctx, "future-pool")
	require.NoError(suite.T(), err)
	require.True(suite.T(), retrievedPool.Active)

	// At height 21, pool should be inactive
	suite.ctx = suite.ctx.WithBlockHeight(21)
	suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})

	retrievedPool, err = suite.keeper.GetRewardPool(suite.ctx, "future-pool")
	require.NoError(suite.T(), err)
	require.False(suite.T(), retrievedPool.Active)
}

// Benchmark tests for production performance validation
func BenchmarkBeginBlock(b *testing.B) {
	suite := new(ModuleTestSuite)
	suite.SetupTest()

	// Create substantial test data
	for i := 0; i < 1000; i++ {
		payment := types.Payment{
			Id:             fmt.Sprintf("payment-%d", i),
			FromAddress:    "cosmos1test1",
			ToAddress:      "cosmos1test2",
			Amount:         sdk.NewCoin("uvita", sdk.NewInt(1000)),
			Fee:            sdk.NewCoin("uvita", sdk.NewInt(10)),
			Status:         types.PaymentStatusPending,
			CreationHeight: 1,
			ExpiryHeight:   int64(i + 10),
		}
		suite.keeper.CreatePayment(suite.ctx, payment)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	}
}

func BenchmarkInvariants(b *testing.B) {
	suite := new(ModuleTestSuite)
	suite.SetupTest()

	// Create comprehensive test data
	for i := 0; i < 100; i++ {
		merchant := types.Merchant{
			Address:            fmt.Sprintf("cosmos1merchant%d", i),
			BusinessName:       fmt.Sprintf("Business %d", i),
			Tier:               types.MerchantTierBronze,
			StakeAmount:        sdk.NewCoin("uvita", sdk.NewInt(1000)),
			TotalVolume:        sdk.NewCoin("uvita", sdk.NewInt(500)),
			RegistrationHeight: int64(i + 1),
		}
		suite.keeper.CreateMerchant(suite.ctx, merchant)

		payment := types.Payment{
			Id:               fmt.Sprintf("payment-%d", i),
			FromAddress:      "cosmos1test1",
			ToAddress:        merchant.Address,
			Amount:           sdk.NewCoin("uvita", sdk.NewInt(100)),
			Fee:              sdk.NewCoin("uvita", sdk.NewInt(1)),
			Status:           types.PaymentStatusCompleted,
			CreationHeight:   int64(i + 1),
			CompletionHeight: int64(i + 2),
		}
		suite.keeper.CreatePayment(suite.ctx, payment)

		vault := types.Vault{
			Id:               fmt.Sprintf("vault-%d", i),
			Owner:            fmt.Sprintf("cosmos1owner%d", i),
			Amount:           sdk.NewCoin("uvita", sdk.NewInt(1000)),
			CreationHeight:   int64(i + 1),
			UnlockHeight:     int64(i + 100),
			RewardMultiplier: sdk.NewDecWithPrec(15, 1),
			Withdrawn:        false,
		}
		suite.keeper.CreateVault(suite.ctx, vault)

		rewardPool := types.RewardPool{
			Id:                 fmt.Sprintf("pool-%d", i),
			MerchantAddress:    merchant.Address,
			TotalRewards:       sdk.NewCoin("uvita", sdk.NewInt(1000)),
			DistributedRewards: sdk.NewCoin("uvita", sdk.NewInt(100)),
			StartHeight:        int64(i + 1),
			EndHeight:          0,
			Active:             true,
		}
		suite.keeper.CreateRewardPool(suite.ctx, rewardPool)
	}

	invariants := keeper.AllInvariants(suite.keeper)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invariants(suite.ctx)
	}
}

func BenchmarkEndBlock(b *testing.B) {
	suite := new(ModuleTestSuite)
	suite.SetupTest()

	// Create test data for end block processing
	for i := 0; i < 50; i++ {
		merchant := types.Merchant{
			Address:            fmt.Sprintf("cosmos1merchant%d", i),
			BusinessName:       fmt.Sprintf("Business %d", i),
			Tier:               types.MerchantTierBronze,
			StakeAmount:        sdk.NewCoin("uvita", sdk.NewInt(1000)),
			TotalVolume:        sdk.NewCoin("uvita", sdk.NewInt(int64(i * 1000))),
			RegistrationHeight: int64(i + 1),
		}
		suite.keeper.CreateMerchant(suite.ctx, merchant)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suite.module.EndBlock(suite.ctx, abci.RequestEndBlock{})
	}
}

func (suite *ModuleTestSuite) TestAppModuleBasic() {
	basic := vitacoin.AppModuleBasic{}

	// Test module name
	require.Equal(suite.T(), types.ModuleName, basic.Name())

	// Test interfaces registration
	encCfg := moduletestutil.MakeTestEncodingConfig(vitacoin.AppModuleBasic{})
	registry := encCfg.InterfaceRegistry
	require.NotPanics(suite.T(), func() {
		basic.RegisterInterfaces(registry)
	})

	// Test default genesis
	genesis := basic.DefaultGenesis(suite.cdc)
	require.NotNil(suite.T(), genesis)

	// Test genesis validation
	err := basic.ValidateGenesis(suite.cdc, nil, genesis)
	require.NoError(suite.T(), err)
}

func (suite *ModuleTestSuite) TestAppModule() {
	// Test module name
	require.Equal(suite.T(), types.ModuleName, suite.module.Name())

	// Test invariants registration - should not panic
	ir := sdk.NewInvariantRegistry()
	require.NotPanics(suite.T(), func() {
		suite.module.RegisterInvariants(ir)
	})

	// Test query routes
	routes := suite.module.LegacyQuerierHandler(nil)
	require.Nil(suite.T(), routes) // Should be nil as we use gRPC

	// Test services registration
	cfg := &moduletestutil.TestConfigurator{}
	require.NotPanics(suite.T(), func() {
		suite.module.RegisterServices(cfg)
	})
}

func (suite *ModuleTestSuite) TestBeginEndBlock() {
	// Setup initial state with a payment that should expire
	payment := types.Payment{
		Id:             "test-payment-1",
		FromAddress:    "cosmos1test1",
		ToAddress:      "cosmos1test2", 
		Amount:         sdk.NewCoin("uvita", sdk.NewInt(1000)),
		Fee:            sdk.NewCoin("uvita", sdk.NewInt(10)),
		Status:         types.PaymentStatusPending,
		CreationHeight: 1,
		ExpiryHeight:   5, // Will expire at height 5
	}

	err := suite.keeper.CreatePayment(suite.ctx, payment)
	require.NoError(suite.T(), err)

	// Test BeginBlock at height 1 - payment should still be pending
	suite.ctx = suite.ctx.WithBlockHeight(1)
	require.NotPanics(suite.T(), func() {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	})

	retrievedPayment, err := suite.keeper.GetPayment(suite.ctx, "test-payment-1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusPending, retrievedPayment.Status)

	// Test BeginBlock at height 6 - payment should be expired
	suite.ctx = suite.ctx.WithBlockHeight(6)
	require.NotPanics(suite.T(), func() {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	})

	retrievedPayment, err = suite.keeper.GetPayment(suite.ctx, "test-payment-1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.PaymentStatusExpired, retrievedPayment.Status)

	// Test EndBlock - should not panic
	require.NotPanics(suite.T(), func() {
		suite.module.EndBlock(suite.ctx, abci.RequestEndBlock{})
	})
}

func (suite *ModuleTestSuite) TestConsensusVersion() {
	// Test consensus version
	version := suite.module.ConsensusVersion()
	require.Equal(suite.T(), uint64(1), version)
}

func (suite *ModuleTestSuite) TestInvariants() {
	// Create test data
	merchant := types.Merchant{
		Address:            "cosmos1merchant",
		BusinessName:       "Test Business",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        sdk.NewCoin("uvita", sdk.NewInt(1000)),
		TotalVolume:        sdk.NewCoin("uvita", sdk.NewInt(500)), // Low volume = Bronze tier
		RegistrationHeight: 1,
	}

	err := suite.keeper.CreateMerchant(suite.ctx, merchant)
	require.NoError(suite.T(), err)

	payment := types.Payment{
		Id:             "test-payment",
		FromAddress:    "cosmos1test1",
		ToAddress:      "cosmos1merchant",
		Amount:         sdk.NewCoin("uvita", sdk.NewInt(100)),
		Fee:            sdk.NewCoin("uvita", sdk.NewInt(1)),
		Status:         types.PaymentStatusCompleted,
		CreationHeight: 1,
		CompletionHeight: 2,
	}

	err = suite.keeper.CreatePayment(suite.ctx, payment)
	require.NoError(suite.T(), err)

	vault := types.Vault{
		Id:               "test-vault",
		Owner:            "cosmos1test1",
		Amount:           sdk.NewCoin("uvita", sdk.NewInt(1000)),
		CreationHeight:   1,
		UnlockHeight:     100,
		RewardMultiplier: sdk.NewDecWithPrec(15, 1), // 1.5x
		Withdrawn:        false,
	}

	err = suite.keeper.CreateVault(suite.ctx, vault)
	require.NoError(suite.T(), err)

	rewardPool := types.RewardPool{
		Id:                "test-pool",
		MerchantAddress:   "cosmos1merchant",
		TotalRewards:      sdk.NewCoin("uvita", sdk.NewInt(1000)),
		DistributedRewards: sdk.NewCoin("uvita", sdk.NewInt(200)),
		StartHeight:       1,
		EndHeight:         0,
		Active:            true,
	}

	err = suite.keeper.CreateRewardPool(suite.ctx, rewardPool)
	require.NoError(suite.T(), err)

	// Test all invariants
	invariants := keeper.AllInvariants(suite.keeper)
	_, broken := invariants(suite.ctx)
	require.False(suite.T(), broken, "Invariants should not be broken with valid data")

	// Test individual invariants
	_, broken = keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Payment invariant should not be broken")

	_, broken = keeper.VaultConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Vault invariant should not be broken")

	_, broken = keeper.RewardPoolConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Reward pool invariant should not be broken")

	_, broken = keeper.MerchantStakeConsistencyInvariant(suite.keeper)(suite.ctx)
	require.False(suite.T(), broken, "Merchant stake invariant should not be broken")
}

func (suite *ModuleTestSuite) TestBrokenInvariants() {
	// Create invalid payment data to break invariants
	invalidPayment := types.Payment{
		Id:               "invalid-payment",
		FromAddress:      "invalid-address", // Invalid bech32 address
		ToAddress:        "cosmos1merchant",
		Amount:           sdk.Coin{}, // Invalid amount
		Fee:              sdk.NewCoin("uvita", sdk.NewInt(-1)), // Negative fee
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   5,
		CompletionHeight: 2, // Completion before creation
	}

	// Manually store invalid data to test invariant detection
	// Note: Normal creation methods would validate, so we test the invariants themselves
	store := suite.ctx.KVStore(suite.keeper.GetStoreKey())
	bz, err := suite.cdc.Marshal(&invalidPayment)
	require.NoError(suite.T(), err)
	store.Set(types.PaymentKey(invalidPayment.Id), bz)

	// Test payment invariant with invalid data
	_, broken := keeper.PaymentConsistencyInvariant(suite.keeper)(suite.ctx)
	require.True(suite.T(), broken, "Payment invariant should be broken with invalid data")

	// Test that all invariants detect the broken state
	_, broken = keeper.AllInvariants(suite.keeper)(suite.ctx)
	require.True(suite.T(), broken, "All invariants should detect broken state")
}

// Test helper functions and edge cases
func (suite *ModuleTestSuite) TestModuleHelpers() {
	// Test merchant tier calculation
	lowVolume := sdk.NewCoin("uvita", sdk.NewInt(500))
	tier := suite.keeper.CalculateMerchantTier(lowVolume)
	require.Equal(suite.T(), types.MerchantTierBronze, tier)

	mediumVolume := sdk.NewCoin("uvita", sdk.NewInt(5000))
	tier = suite.keeper.CalculateMerchantTier(mediumVolume)
	require.Equal(suite.T(), types.MerchantTierSilver, tier)

	highVolume := sdk.NewCoin("uvita", sdk.NewInt(50000))
	tier = suite.keeper.CalculateMerchantTier(highVolume)
	require.Equal(suite.T(), types.MerchantTierGold, tier)
}

// Benchmark tests for module operations
func BenchmarkBeginBlock(b *testing.B) {
	suite := new(ModuleTestSuite)
	suite.SetupTest()

	// Create some test data
	for i := 0; i < 100; i++ {
		payment := types.Payment{
			Id:             fmt.Sprintf("payment-%d", i),
			FromAddress:    "cosmos1test1",
			ToAddress:      "cosmos1test2",
			Amount:         sdk.NewCoin("uvita", sdk.NewInt(1000)),
			Fee:            sdk.NewCoin("uvita", sdk.NewInt(10)),
			Status:         types.PaymentStatusPending,
			CreationHeight: 1,
			ExpiryHeight:   int64(i + 10),
		}
		suite.keeper.CreatePayment(suite.ctx, payment)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suite.module.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	}
}

func BenchmarkInvariants(b *testing.B) {
	suite := new(ModuleTestSuite)
	suite.SetupTest()

	// Create test data
	for i := 0; i < 50; i++ {
		merchant := types.Merchant{
			Address:            fmt.Sprintf("cosmos1merchant%d", i),
			BusinessName:       fmt.Sprintf("Business %d", i),
			Tier:               types.MerchantTierBronze,
			StakeAmount:        sdk.NewCoin("uvita", sdk.NewInt(1000)),
			TotalVolume:        sdk.NewCoin("uvita", sdk.NewInt(500)),
			RegistrationHeight: int64(i + 1),
		}
		suite.keeper.CreateMerchant(suite.ctx, merchant)

		payment := types.Payment{
			Id:               fmt.Sprintf("payment-%d", i),
			FromAddress:      "cosmos1test1",
			ToAddress:        merchant.Address,
			Amount:           sdk.NewCoin("uvita", sdk.NewInt(100)),
			Fee:              sdk.NewCoin("uvita", sdk.NewInt(1)),
			Status:           types.PaymentStatusCompleted,
			CreationHeight:   int64(i + 1),
			CompletionHeight: int64(i + 2),
		}
		suite.keeper.CreatePayment(suite.ctx, payment)
	}

	invariants := keeper.AllInvariants(suite.keeper)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invariants(suite.ctx)
	}
}