package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
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

	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
	msgServer types.MsgServer
	queryServer types.QueryServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	suite.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	// Create context
	suite.ctx = sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Create keeper
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create msg and query servers
	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	// Set default params
	err := suite.keeper.SetParams(suite.ctx, types.DefaultParams())
	require.NoError(suite.T(), err)
}

func (suite *KeeperTestSuite) TestSetGetParams() {
	params := types.DefaultParams()
	params.TransactionFeePercent = math.LegacyNewDecWithPrec(10, 1) // 1.0%

	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	retrievedParams, err := suite.keeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(params, retrievedParams)
}

func (suite *KeeperTestSuite) TestMerchantCRUD() {
	address := "vita1test1merchant"
	
	// Test merchant doesn't exist
	exists, err := suite.keeper.HasMerchant(suite.ctx, address)
	suite.Require().NoError(err)
	suite.Require().False(exists)

	// Create merchant
	merchant := types.Merchant{
		Address:            address,
		BusinessName:       "Test Business",
		Tier:               types.MerchantTierBronze,
		StakeAmount:        math.NewInt(1000000),
		RegistrationHeight: suite.ctx.BlockHeight(),
		IsActive:           true,
		TotalTransactions:  0,
		TotalVolume:        math.ZeroInt(),
	}

	err = suite.keeper.SetMerchant(suite.ctx, merchant)
	suite.Require().NoError(err)

	// Test merchant exists
	exists, err = suite.keeper.HasMerchant(suite.ctx, address)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Get merchant
	retrievedMerchant, err := suite.keeper.GetMerchant(suite.ctx, address)
	suite.Require().NoError(err)
	suite.Require().Equal(merchant.Address, retrievedMerchant.Address)
	suite.Require().Equal(merchant.BusinessName, retrievedMerchant.BusinessName)
	suite.Require().Equal(merchant.Tier, retrievedMerchant.Tier)

	// Get all merchants
	merchants, err := suite.keeper.GetAllMerchants(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(merchants, 1)

	// Delete merchant
	err = suite.keeper.DeleteMerchant(suite.ctx, address)
	suite.Require().NoError(err)

	// Test merchant no longer exists
	exists, err = suite.keeper.HasMerchant(suite.ctx, address)
	suite.Require().NoError(err)
	suite.Require().False(exists)
}

func (suite *KeeperTestSuite) TestPaymentCRUD() {
	paymentID := "payment-1"

	// Test payment doesn't exist
	exists, err := suite.keeper.HasPayment(suite.ctx, paymentID)
	suite.Require().NoError(err)
	suite.Require().False(exists)

	// Create payment
	payment := types.Payment{
		Id:             paymentID,
		FromAddress:    "vita1sender",
		ToAddress:      "vita1merchant",
		Amount:         math.NewInt(1000),
		Status:         types.PaymentStatusPending,
		CreationHeight: suite.ctx.BlockHeight(),
		Memo:           "Test payment",
	}

	err = suite.keeper.SetPayment(suite.ctx, payment)
	suite.Require().NoError(err)

	// Test payment exists
	exists, err = suite.keeper.HasPayment(suite.ctx, paymentID)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Get payment
	retrievedPayment, err := suite.keeper.GetPayment(suite.ctx, paymentID)
	suite.Require().NoError(err)
	suite.Require().Equal(payment.Id, retrievedPayment.Id)
	suite.Require().Equal(payment.FromAddress, retrievedPayment.FromAddress)
	suite.Require().Equal(payment.ToAddress, retrievedPayment.ToAddress)
	suite.Require().Equal(payment.Status, retrievedPayment.Status)

	// Get all payments
	payments, err := suite.keeper.GetAllPayments(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(payments, 1)

	// Delete payment
	err = suite.keeper.DeletePayment(suite.ctx, paymentID)
	suite.Require().NoError(err)

	// Test payment no longer exists
	exists, err = suite.keeper.HasPayment(suite.ctx, paymentID)
	suite.Require().NoError(err)
	suite.Require().False(exists)
}

func (suite *KeeperTestSuite) TestVaultCRUD() {
	vaultID := "vault-1"

	// Test vault doesn't exist
	exists, err := suite.keeper.HasVault(suite.ctx, vaultID)
	suite.Require().NoError(err)
	suite.Require().False(exists)

	// Create vault
	vault := types.Vault{
		Id:               vaultID,
		Owner:            "vita1owner",
		Amount:           math.NewInt(5000),
		LockDuration:     1000,
		CreationHeight:   suite.ctx.BlockHeight(),
		UnlockHeight:     suite.ctx.BlockHeight() + 1000,
		RewardMultiplier: math.LegacyNewDec(2),
	}

	err = suite.keeper.SetVault(suite.ctx, vault)
	suite.Require().NoError(err)

	// Test vault exists
	exists, err = suite.keeper.HasVault(suite.ctx, vaultID)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Get vault
	retrievedVault, err := suite.keeper.GetVault(suite.ctx, vaultID)
	suite.Require().NoError(err)
	suite.Require().Equal(vault.Id, retrievedVault.Id)
	suite.Require().Equal(vault.Owner, retrievedVault.Owner)
	suite.Require().Equal(vault.Amount, retrievedVault.Amount)
	suite.Require().Equal(vault.LockDuration, retrievedVault.LockDuration)

	// Get all vaults
	vaults, err := suite.keeper.GetAllVaults(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(vaults, 1)

	// Delete vault
	err = suite.keeper.DeleteVault(suite.ctx, vaultID)
	suite.Require().NoError(err)

	// Test vault no longer exists
	exists, err = suite.keeper.HasVault(suite.ctx, vaultID)
	suite.Require().NoError(err)
	suite.Require().False(exists)
}

func (suite *KeeperTestSuite) TestRewardPoolCRUD() {
	poolID := "pool-1"

	// Test pool doesn't exist
	exists, err := suite.keeper.HasRewardPool(suite.ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().False(exists)

	// Create reward pool
	pool := types.RewardPool{
		Id:                 poolID,
		MerchantAddress:    "vita1merchant",
		TotalRewards:       math.NewInt(10000),
		DistributedRewards: math.ZeroInt(),
		StartHeight:        suite.ctx.BlockHeight(),
		EndHeight:          suite.ctx.BlockHeight() + 10000,
		IsActive:           true,
	}

	err = suite.keeper.SetRewardPool(suite.ctx, pool)
	suite.Require().NoError(err)

	// Test pool exists
	exists, err = suite.keeper.HasRewardPool(suite.ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().True(exists)

	// Get pool
	retrievedPool, err := suite.keeper.GetRewardPool(suite.ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(pool.Id, retrievedPool.Id)
	suite.Require().Equal(pool.MerchantAddress, retrievedPool.MerchantAddress)
	suite.Require().Equal(pool.TotalRewards, retrievedPool.TotalRewards)

	// Get all pools
	pools, err := suite.keeper.GetAllRewardPools(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(pools, 1)

	// Delete pool
	err = suite.keeper.DeleteRewardPool(suite.ctx, poolID)
	suite.Require().NoError(err)

	// Test pool no longer exists
	exists, err = suite.keeper.HasRewardPool(suite.ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().False(exists)
}

func (suite *KeeperTestSuite) TestGenesisInitAndExport() {
	// Create initial state
	merchant := types.Merchant{
		Address:           "vita1merchant",
		BusinessName:      "Genesis Merchant",
		Tier:              types.MerchantTierGold,
		StakeAmount:       math.NewInt(1000000),
		TotalTransactions: 10,
		TotalVolume:       math.NewInt(500000),
		IsActive:          true,
	}

	payment := types.Payment{
		Id:          "genesis-payment",
		FromAddress: "vita1sender",
		ToAddress:   "vita1merchant",
		Amount:      math.NewInt(1000),
		Status:      types.PaymentStatusCompleted,
	}

	// Create genesis state
	genesisState := &types.GenesisState{
		Params:       types.DefaultParams(),
		MerchantList: []types.Merchant{merchant},
		PaymentList:  []types.Payment{payment},
		VaultList:    []types.Vault{},
		PoolList:     []types.RewardPool{},
	}

	// Initialize genesis
	err := suite.keeper.InitGenesis(suite.ctx, genesisState)
	suite.Require().NoError(err)

	// Verify state was loaded
	retrievedMerchant, err := suite.keeper.GetMerchant(suite.ctx, merchant.Address)
	suite.Require().NoError(err)
	suite.Require().Equal(merchant.BusinessName, retrievedMerchant.BusinessName)

	retrievedPayment, err := suite.keeper.GetPayment(suite.ctx, payment.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(payment.Amount, retrievedPayment.Amount)

	// Export genesis
	exportedState, err := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(exportedState.MerchantList, 1)
	suite.Require().Len(exportedState.PaymentList, 1)
	suite.Require().Equal(merchant.Address, exportedState.MerchantList[0].Address)
	suite.Require().Equal(payment.Id, exportedState.PaymentList[0].Id)
}

func (suite *KeeperTestSuite) TestValidateAuthority() {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Test valid authority
	err := suite.keeper.ValidateAuthority(authority)
	suite.Require().NoError(err)

	// Test invalid authority
	err = suite.keeper.ValidateAuthority("invalid-address")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "unauthorized")
}
