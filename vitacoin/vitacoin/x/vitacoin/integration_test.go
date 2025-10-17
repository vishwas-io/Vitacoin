package vitacoin_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/log"
	
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin"
	"github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// IntegrationTestSuite provides a comprehensive integration testing suite for the vitacoin module
// This suite tests the module with full keeper functionality and state management
type IntegrationTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    keeper.Keeper
	msgSrv    types.MsgServer
	queryServ types.QueryServer
	
	bankKeeper    *MockBankKeeper
	accountKeeper MockAccountKeeper
	
	// Test accounts
	testAccounts map[string]sdk.AccAddress
}

// MockBankKeeper provides a production-grade mock bank keeper for integration testing
// It simulates real bank keeper behavior including balance tracking and transfers
type MockBankKeeper struct {
	balances map[string]sdk.Coins
	locked   map[string]sdk.Coins // Simulates module account locked funds
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]sdk.Coins),
		locked:   make(map[string]sdk.Coins),
	}
}

// MintCoins simulates minting coins to a module account
func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	if !amounts.IsValid() {
		return fmt.Errorf("invalid coins: %s", amounts)
	}
	
	addr := authtypes.NewModuleAddress(moduleName).String()
	if m.balances[addr] == nil {
		m.balances[addr] = sdk.NewCoins()
	}
	m.balances[addr] = m.balances[addr].Add(amounts...)
	return nil
}

// BurnCoins simulates burning coins from a module account
func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	if !amounts.IsValid() {
		return fmt.Errorf("invalid coins: %s", amounts)
	}
	
	addr := authtypes.NewModuleAddress(moduleName).String()
	if m.balances[addr] == nil || m.balances[addr].IsAllLT(amounts) {
		return fmt.Errorf("insufficient balance in module %s", moduleName)
	}
	m.balances[addr] = m.balances[addr].Sub(amounts...)
	return nil
}

// SendCoinsFromModuleToAccount simulates sending coins from a module to an account
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if !amt.IsValid() {
		return fmt.Errorf("invalid coins: %s", amt)
	}
	
	senderAddr := authtypes.NewModuleAddress(senderModule).String()
	if m.balances[senderAddr] == nil || m.balances[senderAddr].IsAllLT(amt) {
		return fmt.Errorf("insufficient balance in module %s: has %s, need %s", 
			senderModule, m.balances[senderAddr], amt)
	}
	
	m.balances[senderAddr] = m.balances[senderAddr].Sub(amt...)
	if m.balances[recipientAddr.String()] == nil {
		m.balances[recipientAddr.String()] = sdk.NewCoins()
	}
	m.balances[recipientAddr.String()] = m.balances[recipientAddr.String()].Add(amt...)
	return nil
}

// SendCoinsFromAccountToModule simulates sending coins from an account to a module
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if !amt.IsValid() {
		return fmt.Errorf("invalid coins: %s", amt)
	}
	
	recipientAddr := authtypes.NewModuleAddress(recipientModule).String()
	if m.balances[senderAddr.String()] == nil || m.balances[senderAddr.String()].IsAllLT(amt) {
		return fmt.Errorf("insufficient balance for account %s: has %s, need %s", 
			senderAddr.String(), m.balances[senderAddr.String()], amt)
	}
	
	m.balances[senderAddr.String()] = m.balances[senderAddr.String()].Sub(amt...)
	if m.balances[recipientAddr] == nil {
		m.balances[recipientAddr] = sdk.NewCoins()
	}
	m.balances[recipientAddr] = m.balances[recipientAddr].Add(amt...)
	return nil
}

// SendCoinsFromModuleToModule simulates sending coins between modules
func (m *MockBankKeeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if !amt.IsValid() {
		return fmt.Errorf("invalid coins: %s", amt)
	}
	
	senderAddr := authtypes.NewModuleAddress(senderModule).String()
	recipientAddr := authtypes.NewModuleAddress(recipientModule).String()
	
	if m.balances[senderAddr] == nil || m.balances[senderAddr].IsAllLT(amt) {
		return fmt.Errorf("insufficient balance in module %s: has %s, need %s", 
			senderModule, m.balances[senderAddr], amt)
	}
	
	m.balances[senderAddr] = m.balances[senderAddr].Sub(amt...)
	if m.balances[recipientAddr] == nil {
		m.balances[recipientAddr] = sdk.NewCoins()
	}
	m.balances[recipientAddr] = m.balances[recipientAddr].Add(amt...)
	return nil
}

// GetAllBalances returns all balances for an account
func (m *MockBankKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if m.balances[addr.String()] == nil {
		return sdk.NewCoins()
	}
	return m.balances[addr.String()]
}

// GetBalance returns the balance of a specific denom for an account
func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balances := m.GetAllBalances(ctx, addr)
	return balances.AmountOf(denom)
}

// SetBalance sets the balance for an account (helper for tests)
func (m *MockBankKeeper) SetBalance(addr string, coins sdk.Coins) {
	m.balances[addr] = coins
}

// SpendableCoins returns spendable coins (same as balance in this mock)
func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return m.GetAllBalances(ctx, addr)
}

// MockAccountKeeper provides a production-grade mock account keeper
type MockAccountKeeper struct {
	accounts map[string]sdk.AccountI
}

func NewMockAccountKeeper() MockAccountKeeper {
	return MockAccountKeeper{
		accounts: make(map[string]sdk.AccountI),
	}
}

// GetModuleAddress returns the address for a module
func (m MockAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return authtypes.NewModuleAddress(moduleName)
}

// GetAccount returns an account
func (m MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.accounts[addr.String()]
}

// SetupTest initializes a production-grade test environment with full keeper setup
func (suite *IntegrationTestSuite) SetupTest() {
	// Create test keys for state storage
	keys := storetypes.NewKVStoreKeys(types.StoreKey)
	tkeys := storetypes.NewTransientStoreKeys(types.TransientStoreKey)
	
	// Create test context with proper block header
	ctx := testutil.DefaultContextWithDB(suite.T(), keys[types.StoreKey], tkeys[types.TransientStoreKey])
	suite.ctx = ctx.Ctx.WithBlockHeight(1).WithBlockTime(time.Now())
	
	// Create encoding config with all required codecs
	encCfg := moduletestutil.MakeTestEncodingConfig(vitacoin.AppModuleBasic{})
	
	// Create mock keepers with full functionality
	suite.accountKeeper = NewMockAccountKeeper()
	suite.bankKeeper = NewMockBankKeeper()
	
	// Create the vitacoin keeper with all dependencies
	suite.keeper = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		log.NewNopLogger(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		suite.accountKeeper,
		suite.bankKeeper,
	)
	
	// Initialize module parameters with production defaults
	err := suite.keeper.SetParams(suite.ctx, types.DefaultParams())
	require.NoError(suite.T(), err)
	
	// Create message and query servers
	suite.msgSrv = keeper.NewMsgServerImpl(suite.keeper)
	suite.queryServ = keeper.NewQueryServerImpl(suite.keeper)
	
	// Initialize test accounts
	suite.testAccounts = make(map[string]sdk.AccAddress)
	suite.testAccounts["merchant1"] = sdk.AccAddress([]byte("merchant1___________"))
	suite.testAccounts["merchant2"] = sdk.AccAddress([]byte("merchant2___________"))
	suite.testAccounts["customer1"] = sdk.AccAddress([]byte("customer1___________"))
	suite.testAccounts["customer2"] = sdk.AccAddress([]byte("customer2___________"))
	suite.testAccounts["vault_user"] = sdk.AccAddress([]byte("vault_user__________"))
	suite.testAccounts["admin"] = authtypes.NewModuleAddress(govtypes.ModuleName)
	
	// Fund all test accounts with initial balances
	for name, addr := range suite.testAccounts {
		if name != "admin" {
			coins := sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(10000000000000000000000))) // 10,000 VITA
			suite.bankKeeper.SetBalance(addr.String(), coins)
		}
	}
	
	// Fund admin account for governance operations
	adminCoins := sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(100000000000000000000000))) // 100,000 VITA
	suite.bankKeeper.SetBalance(suite.testAccounts["admin"].String(), adminCoins)
}

// TestIntegrationTestSuite runs the comprehensive integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// TestMerchantLifecycle tests the complete merchant lifecycle with all state transitions
func (suite *IntegrationTestSuite) TestMerchantLifecycle() {
	addr := suite.testAccounts["merchant1"]
	
	// Test 1: Register a new merchant
	registerMsg := &types.MsgRegisterMerchant{
		Creator:      addr.String(),
		BusinessName: "Premium Coffee Shop",
		Category:     "Food & Beverage",
		StakeAmount:  "10000000000000000000000", // 10,000 VITA (Silver tier)
	}

	_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.NoError(suite.T(), err, "merchant registration should succeed")

	// Verify merchant state
	merchant, found := suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.True(suite.T(), found, "merchant should exist in state")
	require.Equal(suite.T(), "Premium Coffee Shop", merchant.BusinessName)
	require.Equal(suite.T(), types.MerchantTier_SILVER, merchant.Tier)
	require.True(suite.T(), merchant.IsActive)
	require.Equal(suite.T(), "10000000000000000000000", merchant.StakeAmount)
	require.Greater(suite.T(), merchant.CreatedAt, uint64(0))

	// Test 2: Update merchant to Gold tier
	updateMsg := &types.MsgUpdateMerchant{
		Creator:      addr.String(),
		BusinessName: "Premium Coffee Shop - Downtown",
		Category:     "Food & Beverage - Premium",
		StakeAmount:  "100000000000000000000000", // 100,000 VITA (Gold tier)
	}

	_, err = suite.msgSrv.UpdateMerchant(suite.ctx, updateMsg)
	require.NoError(suite.T(), err, "merchant update should succeed")

	// Verify tier upgrade
	merchant, found = suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), "Premium Coffee Shop - Downtown", merchant.BusinessName)
	require.Equal(suite.T(), types.MerchantTier_GOLD, merchant.Tier)
	require.Equal(suite.T(), "100000000000000000000000", merchant.StakeAmount)

	// Test 3: Verify discount calculation for Gold tier
	discount := keeper.CalculateFeeDiscount(merchant.Tier)
	require.Equal(suite.T(), uint64(5000), discount, "Gold tier should have 50% discount")

	// Test 4: Deactivate merchant
	deactivateMsg := &types.MsgDeactivateMerchant{
		Creator: addr.String(),
	}

	_, err = suite.msgSrv.DeactivateMerchant(suite.ctx, deactivateMsg)
	require.NoError(suite.T(), err, "merchant deactivation should succeed")

	// Verify deactivation
	merchant, found = suite.keeper.GetMerchant(suite.ctx, addr.String())
	require.True(suite.T(), found)
	require.False(suite.T(), merchant.IsActive, "merchant should be inactive")

	// Test 5: Verify inactive merchant cannot accept payments
	customerAddr := suite.testAccounts["customer1"]
	createPaymentMsg := &types.MsgCreatePayment{
		Creator:    customerAddr.String(),
		MerchantId: addr.String(),
		Amount:     "1000000000000000000", // 1 VITA
		Currency:   "VITA",
		PaymentId:  "payment_inactive_test",
		Memo:       "Should fail for inactive merchant",
	}
	
	_, err = suite.msgSrv.CreatePayment(suite.ctx, createPaymentMsg)
	require.Error(suite.T(), err, "should not allow payment to inactive merchant")
	require.Contains(suite.T(), err.Error(), "merchant is not active")
}

// TestPaymentFlow tests the complete payment lifecycle including fees and status transitions
func (suite *IntegrationTestSuite) TestPaymentFlow() {
	merchantAddr := suite.testAccounts["merchant1"]
	customerAddr := suite.testAccounts["customer1"]

	// Register merchant first
	registerMsg := &types.MsgRegisterMerchant{
		Creator:      merchantAddr.String(),
		BusinessName: "Test E-commerce Store",
		Category:     "E-commerce",
		StakeAmount:  "10000000000000000000000", // 10,000 VITA (Silver tier)
	}
	_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.NoError(suite.T(), err)

	// Get initial balances
	initialMerchantBalance := suite.bankKeeper.GetAllBalances(suite.ctx, merchantAddr)
	initialCustomerBalance := suite.bankKeeper.GetAllBalances(suite.ctx, customerAddr)

	// Test 1: Create a payment
	paymentID := fmt.Sprintf("payment_%d", time.Now().Unix())
	paymentAmount := sdkmath.NewInt(1000000000000000000) // 1 VITA
	
	createPaymentMsg := &types.MsgCreatePayment{
		Creator:    customerAddr.String(),
		MerchantId: merchantAddr.String(),
		Amount:     paymentAmount.String(),
		Currency:   "VITA",
		PaymentId:  paymentID,
		Memo:       "Integration test payment",
	}

	_, err = suite.msgSrv.CreatePayment(suite.ctx, createPaymentMsg)
	require.NoError(suite.T(), err, "payment creation should succeed")

	// Verify payment state
	payment, found := suite.keeper.GetPayment(suite.ctx, paymentID)
	require.True(suite.T(), found, "payment should exist")
	require.Equal(suite.T(), types.PaymentStatus_PENDING, payment.Status)
	require.Equal(suite.T(), customerAddr.String(), payment.CustomerId)
	require.Equal(suite.T(), merchantAddr.String(), payment.MerchantId)
	require.Equal(suite.T(), paymentAmount.String(), payment.Amount)

	// Test 2: Complete the payment
	completePaymentMsg := &types.MsgCompletePayment{
		Creator:   merchantAddr.String(),
		PaymentId: paymentID,
	}

	_, err = suite.msgSrv.CompletePayment(suite.ctx, completePaymentMsg)
	require.NoError(suite.T(), err, "payment completion should succeed")

	// Verify payment status update
	payment, found = suite.keeper.GetPayment(suite.ctx, paymentID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.PaymentStatus_COMPLETED, payment.Status)
	require.Greater(suite.T(), payment.CompletedAt, uint64(0))

	// Test 3: Verify fee calculation (Silver tier has 25% discount)
	params, _ := suite.keeper.GetParams(suite.ctx)
	baseFee, _ := sdkmath.NewIntFromString(params.TransactionFee)
	merchant, _ := suite.keeper.GetMerchant(suite.ctx, merchantAddr.String())
	actualFee := keeper.CalculateTransactionFee(paymentAmount, baseFee, merchant.Tier)
	require.Greater(suite.T(), actualFee.Int64(), int64(0), "fee should be calculated")

	// Test 4: Refund the payment
	refundPaymentMsg := &types.MsgRefundPayment{
		Creator:   merchantAddr.String(),
		PaymentId: paymentID,
		Reason:    "Customer requested refund - integration test",
	}

	_, err = suite.msgSrv.RefundPayment(suite.ctx, refundPaymentMsg)
	require.NoError(suite.T(), err, "payment refund should succeed")

	// Verify refund state
	payment, found = suite.keeper.GetPayment(suite.ctx, paymentID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.PaymentStatus_REFUNDED, payment.Status)
	require.Equal(suite.T(), "Customer requested refund - integration test", payment.RefundReason)

	// Test 5: Verify cannot complete refunded payment
	_, err = suite.msgSrv.CompletePayment(suite.ctx, completePaymentMsg)
	require.Error(suite.T(), err, "should not complete already refunded payment")
	require.Contains(suite.T(), err.Error(), "payment is not in pending status")
}

// TestVaultOperations tests vault creation, deposits, withdrawals, and reward calculations
func (suite *IntegrationTestSuite) TestVaultOperations() {
	userAddr := suite.testAccounts["vault_user"]

	// Test 1: Create a vault with lock duration
	lockDuration := uint64(1000000) // ~1 week in blocks
	initialAmount := sdkmath.NewInt(100000000000000000000) // 100 VITA
	
	createVaultMsg := &types.MsgCreateVault{
		Creator:      userAddr.String(),
		Amount:       initialAmount.String(),
		LockDuration: lockDuration,
	}

	resp, err := suite.msgSrv.CreateVault(suite.ctx, createVaultMsg)
	require.NoError(suite.T(), err, "vault creation should succeed")
	vaultID := resp.VaultId
	require.NotEmpty(suite.T(), vaultID, "vault ID should be generated")

	// Verify vault state
	vault, found := suite.keeper.GetVault(suite.ctx, vaultID)
	require.True(suite.T(), found, "vault should exist")
	require.Equal(suite.T(), userAddr.String(), vault.Owner)
	require.Equal(suite.T(), initialAmount.String(), vault.LockedAmount)
	require.Greater(suite.T(), vault.UnlockHeight, uint64(suite.ctx.BlockHeight()))
	require.Greater(suite.T(), vault.RewardMultiplier, "0")

	// Test 2: Deposit additional funds
	depositAmount := sdkmath.NewInt(50000000000000000000) // 50 VITA
	depositMsg := &types.MsgDepositToVault{
		Creator: userAddr.String(),
		VaultId: vaultID,
		Amount:  depositAmount.String(),
	}

	_, err = suite.msgSrv.DepositToVault(suite.ctx, depositMsg)
	require.NoError(suite.T(), err, "vault deposit should succeed")

	// Verify updated balance
	vault, found = suite.keeper.GetVault(suite.ctx, vaultID)
	require.True(suite.T(), found)
	expectedTotal := initialAmount.Add(depositAmount)
	actualTotal, ok := sdkmath.NewIntFromString(vault.LockedAmount)
	require.True(suite.T(), ok)
	require.True(suite.T(), expectedTotal.Equal(actualTotal), "vault balance should be updated")

	// Test 3: Try to withdraw before unlock (should fail)
	withdrawMsg := &types.MsgWithdrawFromVault{
		Creator: userAddr.String(),
		VaultId: vaultID,
		Amount:  "10000000000000000000", // 10 VITA
	}

	_, err = suite.msgSrv.WithdrawFromVault(suite.ctx, withdrawMsg)
	require.Error(suite.T(), err, "should not allow withdrawal before unlock")
	require.Contains(suite.T(), err.Error(), "vault is still locked")

	// Test 4: Fast forward past unlock height
	vault, _ = suite.keeper.GetVault(suite.ctx, vaultID)
	suite.ctx = suite.ctx.WithBlockHeight(int64(vault.UnlockHeight) + 1)

	// Test 5: Withdraw after unlock
	withdrawAmount := sdkmath.NewInt(50000000000000000000) // 50 VITA
	withdrawMsg.Amount = withdrawAmount.String()

	_, err = suite.msgSrv.WithdrawFromVault(suite.ctx, withdrawMsg)
	require.NoError(suite.T(), err, "withdrawal should succeed after unlock")

	// Verify remaining balance
	vault, found = suite.keeper.GetVault(suite.ctx, vaultID)
	require.True(suite.T(), found)
	expectedRemaining := expectedTotal.Sub(withdrawAmount)
	actualRemaining, ok := sdkmath.NewIntFromString(vault.LockedAmount)
	require.True(suite.T(), ok)
	require.True(suite.T(), expectedRemaining.Equal(actualRemaining), "vault balance should be updated after withdrawal")

	// Test 6: Verify reward calculation
	rewardMultiplier, ok := sdkmath.NewIntFromString(vault.RewardMultiplier)
	require.True(suite.T(), ok)
	require.Greater(suite.T(), rewardMultiplier.Int64(), int64(1000000000000000000), "reward multiplier should be set based on lock duration")
}

// TestRewardDistribution tests reward pool creation and distribution with proper validation
func (suite *IntegrationTestSuite) TestRewardDistribution() {
	adminAddr := suite.testAccounts["admin"]
	
	// Test 1: Create a reward pool
	totalRewards := sdkmath.NewInt(5000000000000000000000) // 5,000 VITA
	createPoolMsg := &types.MsgCreateRewardPool{
		Creator:             adminAddr.String(),
		TotalRewards:        totalRewards.String(),
		DistributionPeriod:  86400, // 1 day
		EligibilityCriteria: "Active merchants with > 10 transactions in the last 30 days",
	}

	resp, err := suite.msgSrv.CreateRewardPool(suite.ctx, createPoolMsg)
	require.NoError(suite.T(), err, "reward pool creation should succeed")
	poolID := resp.PoolId
	require.NotEmpty(suite.T(), poolID)

	// Verify pool state
	pool, found := suite.keeper.GetRewardPool(suite.ctx, poolID)
	require.True(suite.T(), found, "reward pool should exist")
	require.Equal(suite.T(), totalRewards.String(), pool.TotalRewards)
	require.Equal(suite.T(), totalRewards.String(), pool.RemainingRewards)
	require.Equal(suite.T(), uint64(86400), pool.DistributionPeriod)

	// Test 2: Prepare recipients
	recipient1 := suite.testAccounts["merchant1"]
	recipient2 := suite.testAccounts["merchant2"]
	
	// Register recipients as merchants
	for i, addr := range []sdk.AccAddress{recipient1, recipient2} {
		registerMsg := &types.MsgRegisterMerchant{
			Creator:      addr.String(),
			BusinessName: fmt.Sprintf("Reward Recipient %d", i+1),
			Category:     "Retail",
			StakeAmount:  "1000000000000000000000", // 1,000 VITA (Bronze)
		}
		_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
		require.NoError(suite.T(), err)
	}

	// Test 3: Distribute rewards
	amount1 := sdkmath.NewInt(3000000000000000000000) // 3,000 VITA
	amount2 := sdkmath.NewInt(2000000000000000000000) // 2,000 VITA
	
	distributeMsg := &types.MsgDistributeRewards{
		Creator:    adminAddr.String(),
		PoolId:     poolID,
		Recipients: []string{recipient1.String(), recipient2.String()},
		Amounts:    []string{amount1.String(), amount2.String()},
	}

	_, err = suite.msgSrv.DistributeRewards(suite.ctx, distributeMsg)
	require.NoError(suite.T(), err, "reward distribution should succeed")

	// Verify pool depletion
	pool, found = suite.keeper.GetRewardPool(suite.ctx, poolID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), "0", pool.RemainingRewards, "all rewards should be distributed")

	// Test 4: Verify cannot over-distribute
	overDistributeMsg := &types.MsgDistributeRewards{
		Creator:    adminAddr.String(),
		PoolId:     poolID,
		Recipients: []string{recipient1.String()},
		Amounts:    []string{"1000000000000000000"}, // 1 VITA
	}

	_, err = suite.msgSrv.DistributeRewards(suite.ctx, overDistributeMsg)
	require.Error(suite.T(), err, "should not allow over-distribution")
	require.Contains(suite.T(), err.Error(), "insufficient rewards")
}

// TestGovernanceIntegration tests governance-based parameter updates
func (suite *IntegrationTestSuite) TestGovernanceIntegration() {
	govAddr := suite.testAccounts["admin"]

	// Test 1: Get current params
	paramsResp, err := suite.queryServ.Params(suite.ctx, &types.QueryParamsRequest{})
	require.NoError(suite.T(), err)
	oldParams := paramsResp.Params

	// Test 2: Update params via governance
	newParams := types.DefaultParams()
	newParams.TransactionFee = "500"                        // 0.05%
	newParams.MinMerchantStake = "2000000000000000000000"  // 2,000 VITA
	newParams.MerchantRegistrationFee = "500000000000000000000" // 500 VITA
	
	updateMsg := &types.MsgUpdateParams{
		Authority: govAddr.String(),
		Params:    newParams,
	}

	_, err = suite.msgSrv.UpdateParams(suite.ctx, updateMsg)
	require.NoError(suite.T(), err, "param update should succeed")

	// Verify params were updated
	updatedParamsResp, err := suite.queryServ.Params(suite.ctx, &types.QueryParamsRequest{})
	require.NoError(suite.T(), err)
	updatedParams := updatedParamsResp.Params
	
	require.NotEqual(suite.T(), oldParams.TransactionFee, updatedParams.TransactionFee)
	require.Equal(suite.T(), "500", updatedParams.TransactionFee)
	require.Equal(suite.T(), "2000000000000000000000", updatedParams.MinMerchantStake)

	// Test 3: Verify non-governance address cannot update params
	nonGovAddr := suite.testAccounts["merchant1"]
	invalidUpdateMsg := &types.MsgUpdateParams{
		Authority: nonGovAddr.String(),
		Params:    newParams,
	}

	_, err = suite.msgSrv.UpdateParams(suite.ctx, invalidUpdateMsg)
	require.Error(suite.T(), err, "non-governance address should not update params")
	require.Contains(suite.T(), err.Error(), "expected gov account as only signer")
}

// TestQueryEndpoints tests all query endpoints with comprehensive data
func (suite *IntegrationTestSuite) TestQueryEndpoints() {
	// Setup test data
	merchantAddr := suite.testAccounts["merchant1"]
	customerAddr := suite.testAccounts["customer1"]

	// Register merchant
	registerMsg := &types.MsgRegisterMerchant{
		Creator:      merchantAddr.String(),
		BusinessName: "Query Test Merchant",
		Category:     "Retail",
		StakeAmount:  "10000000000000000000000",
	}
	_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.NoError(suite.T(), err)

	// Create payment
	paymentID := "query_test_payment"
	createPaymentMsg := &types.MsgCreatePayment{
		Creator:    customerAddr.String(),
		MerchantId: merchantAddr.String(),
		Amount:     "1000000000000000000",
		Currency:   "VITA",
		PaymentId:  paymentID,
		Memo:       "Query test",
	}
	_, err = suite.msgSrv.CreatePayment(suite.ctx, createPaymentMsg)
	require.NoError(suite.T(), err)

	// Test 1: Query Params
	paramsResp, err := suite.queryServ.Params(suite.ctx, &types.QueryParamsRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), paramsResp.Params)

	// Test 2: Query Merchant
	merchantResp, err := suite.queryServ.Merchant(suite.ctx, &types.QueryMerchantRequest{
		MerchantId: merchantAddr.String(),
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Query Test Merchant", merchantResp.Merchant.BusinessName)

	// Test 3: Query AllMerchants
	allMerchantsResp, err := suite.queryServ.AllMerchants(suite.ctx, &types.QueryAllMerchantsRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(allMerchantsResp.Merchants), 1)

	// Test 4: Query Payment
	paymentResp, err := suite.queryServ.Payment(suite.ctx, &types.QueryPaymentRequest{
		PaymentId: paymentID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), paymentID, paymentResp.Payment.PaymentId)

	// Test 5: Query AllPayments
	allPaymentsResp, err := suite.queryServ.AllPayments(suite.ctx, &types.QueryAllPaymentsRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(allPaymentsResp.Payments), 1)

	// Test 6: Query non-existent entities (should error)
	_, err = suite.queryServ.Merchant(suite.ctx, &types.QueryMerchantRequest{
		MerchantId: "non_existent_merchant",
	})
	require.Error(suite.T(), err)

	_, err = suite.queryServ.Payment(suite.ctx, &types.QueryPaymentRequest{
		PaymentId: "non_existent_payment",
	})
	require.Error(suite.T(), err)
}

// TestBlockLifecycle tests BeginBlock and EndBlock operations
func (suite *IntegrationTestSuite) TestBlockLifecycle() {
	// Create test merchants
	merchant1 := suite.testAccounts["merchant1"]
	merchant2 := suite.testAccounts["merchant2"]

	for i, addr := range []sdk.AccAddress{merchant1, merchant2} {
		registerMsg := &types.MsgRegisterMerchant{
			Creator:      addr.String(),
			BusinessName: fmt.Sprintf("Lifecycle Merchant %d", i+1),
			Category:     "Testing",
			StakeAmount:  "10000000000000000000000",
		}
		_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
		require.NoError(suite.T(), err)
	}

	// Execute BeginBlocker
	err := suite.keeper.BeginBlocker(suite.ctx)
	require.NoError(suite.T(), err, "BeginBlocker should execute successfully")

	// Verify state is consistent after BeginBlocker
	merchant, found := suite.keeper.GetMerchant(suite.ctx, merchant1.String())
	require.True(suite.T(), found)
	require.True(suite.T(), merchant.IsActive)

	// Execute EndBlocker
	err = suite.keeper.EndBlocker(suite.ctx)
	require.NoError(suite.T(), err, "EndBlocker should execute successfully")

	// Verify state is consistent after EndBlocker
	merchant, found = suite.keeper.GetMerchant(suite.ctx, merchant1.String())
	require.True(suite.T(), found)
	require.True(suite.T(), merchant.IsActive)
}

// TestConcurrentOperations tests handling of concurrent/sequential operations
func (suite *IntegrationTestSuite) TestConcurrentOperations() {
	// Create multiple merchants sequentially
	numMerchants := 5
	merchantAddrs := make([]sdk.AccAddress, numMerchants)

	for i := 0; i < numMerchants; i++ {
		addr := sdk.AccAddress([]byte(fmt.Sprintf("concurrent_merch_%02d_", i)))
		merchantAddrs[i] = addr
		
		// Fund account
		coins := sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(10000000000000000000000)))
		suite.bankKeeper.SetBalance(addr.String(), coins)
		
		// Register merchant
		registerMsg := &types.MsgRegisterMerchant{
			Creator:      addr.String(),
			BusinessName: fmt.Sprintf("Concurrent Merchant %d", i),
			Category:     "Testing",
			StakeAmount:  "1000000000000000000000",
		}
		_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
		require.NoError(suite.T(), err)
	}

	// Verify all merchants were created
	allMerchantsResp, err := suite.queryServ.AllMerchants(suite.ctx, &types.QueryAllMerchantsRequest{})
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(allMerchantsResp.Merchants), numMerchants)

	// Update all merchants
	for i, addr := range merchantAddrs {
		updateMsg := &types.MsgUpdateMerchant{
			Creator:      addr.String(),
			BusinessName: fmt.Sprintf("Updated Concurrent Merchant %d", i),
			Category:     "Updated Testing",
			StakeAmount:  "2000000000000000000000",
		}
		_, err := suite.msgSrv.UpdateMerchant(suite.ctx, updateMsg)
		require.NoError(suite.T(), err)
	}

	// Verify all updates
	for i, addr := range merchantAddrs {
		merchant, found := suite.keeper.GetMerchant(suite.ctx, addr.String())
		require.True(suite.T(), found)
		require.Equal(suite.T(), fmt.Sprintf("Updated Concurrent Merchant %d", i), merchant.BusinessName)
	}
}

// TestErrorHandling tests comprehensive error scenarios
func (suite *IntegrationTestSuite) TestErrorHandling() {
	addr := suite.testAccounts["merchant1"]

	// Test 1: Insufficient funds for registration
	poorAddr := sdk.AccAddress([]byte("poor_merchant_______"))
	suite.bankKeeper.SetBalance(poorAddr.String(), sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(100))))
	
	registerMsg := &types.MsgRegisterMerchant{
		Creator:      poorAddr.String(),
		BusinessName: "Poor Merchant",
		Category:     "Testing",
		StakeAmount:  "10000000000000000000000",
	}
	_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.Error(suite.T(), err)

	// Test 2: Duplicate merchant registration
	registerMsg = &types.MsgRegisterMerchant{
		Creator:      addr.String(),
		BusinessName: "Test Merchant",
		Category:     "Testing",
		StakeAmount:  "1000000000000000000000",
	}
	_, err = suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.NoError(suite.T(), err)

	_, err = suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "merchant already exists")

	// Test 3: Invalid payment to inactive merchant
	deactivateMsg := &types.MsgDeactivateMerchant{
		Creator: addr.String(),
	}
	_, err = suite.msgSrv.DeactivateMerchant(suite.ctx, deactivateMsg)
	require.NoError(suite.T(), err)

	customerAddr := suite.testAccounts["customer1"]
	createPaymentMsg := &types.MsgCreatePayment{
		Creator:    customerAddr.String(),
		MerchantId: addr.String(),
		Amount:     "1000000000000000000",
		Currency:   "VITA",
		PaymentId:  "invalid_payment",
		Memo:       "Should fail",
	}
	_, err = suite.msgSrv.CreatePayment(suite.ctx, createPaymentMsg)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "not active")

	// Test 4: Invalid vault operations
	userAddr := suite.testAccounts["vault_user"]
	createVaultMsg := &types.MsgCreateVault{
		Creator:      userAddr.String(),
		Amount:       "100000000000000000000",
		LockDuration: 1000000,
	}
	resp, err := suite.msgSrv.CreateVault(suite.ctx, createVaultMsg)
	require.NoError(suite.T(), err)

	// Try to withdraw before unlock
	withdrawMsg := &types.MsgWithdrawFromVault{
		Creator: userAddr.String(),
		VaultId: resp.VaultId,
		Amount:  "10000000000000000000",
	}
	_, err = suite.msgSrv.WithdrawFromVault(suite.ctx, withdrawMsg)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "locked")
}

// TestInvariants tests that invariants hold across operations
func (suite *IntegrationTestSuite) TestInvariants() {
	// This test ensures that module invariants are maintained
	// Invariants are checked in the keeper's invariants.go file

	// Create multiple merchants and payments
	for i := 0; i < 3; i++ {
		addr := sdk.AccAddress([]byte(fmt.Sprintf("invariant_merch%02d__", i)))
		coins := sdk.NewCoins(sdk.NewCoin("avita", sdkmath.NewInt(10000000000000000000000)))
		suite.bankKeeper.SetBalance(addr.String(), coins)

		registerMsg := &types.MsgRegisterMerchant{
			Creator:      addr.String(),
			BusinessName: fmt.Sprintf("Invariant Merchant %d", i),
			Category:     "Testing",
			StakeAmount:  "1000000000000000000000",
		}
		_, err := suite.msgSrv.RegisterMerchant(suite.ctx, registerMsg)
		require.NoError(suite.T(), err)
	}

	// Run BeginBlock and EndBlock to trigger invariant checks
	err := suite.keeper.BeginBlocker(suite.ctx)
	require.NoError(suite.T(), err)
	
	err = suite.keeper.EndBlocker(suite.ctx)
	require.NoError(suite.T(), err)

	// Invariants should pass without panics
	// If invariants fail, the keeper would panic in production
}
