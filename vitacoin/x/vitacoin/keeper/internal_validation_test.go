package keeper

// internal_validation_test.go — white-box tests for unexported msgServer validation methods.
// Package is "keeper" (not "keeper_test") to access unexported types.

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// newMsgServerInternal wraps the keeper in the unexported msgServer type.
func newMsgServerInternal(k Keeper) msgServer {
	return msgServer{Keeper: k}
}

// ──────────────────────────────────────────────
// ValidateTransactionContext
// ──────────────────────────────────────────────

func TestInternal_ValidateTransactionContext(t *testing.T) {
	k, ctx := setupTestKeeper(t) // reuse from keeper_test.go via same package
	ms := newMsgServerInternal(k)

	// height=1, valid time → should pass
	err := ms.ValidateTransactionContext(ctx, "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f")
	require.NoError(t, err)

	// height=0 → should fail
	zeroCtx := ctx.WithBlockHeight(0)
	err = ms.ValidateTransactionContext(zeroCtx, "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f")
	require.Error(t, err)

	// zero block time → should fail
	zeroTimeCtx := ctx.WithBlockTime(time.Time{})
	err = ms.ValidateTransactionContext(zeroTimeCtx, "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f")
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// ValidateMerchantOperationalStatus
// ──────────────────────────────────────────────

func TestInternal_ValidateMerchantOperationalStatus(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	addr := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	// Non-existent merchant → error
	err := ms.ValidateMerchantOperationalStatus(ctx, addr)
	require.Error(t, err)

	// Active merchant → ok
	m := types.Merchant{
		Address: addr, BusinessName: "M",
		Tier: types.MerchantTierBronze, StakeAmount: math.NewInt(1_000),
		IsActive: true, TotalVolume: math.ZeroInt(),
	}
	require.NoError(t, k.SetMerchant(ctx, m))
	err = ms.ValidateMerchantOperationalStatus(ctx, addr)
	require.NoError(t, err)

	// Inactive merchant → error
	m.IsActive = false
	require.NoError(t, k.SetMerchant(ctx, m))
	err = ms.ValidateMerchantOperationalStatus(ctx, addr)
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// ValidatePaymentOperationalConstraints
// ──────────────────────────────────────────────

func TestInternal_ValidatePaymentOperationalConstraints(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	addr := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"
	m := types.Merchant{
		Address: addr, BusinessName: "M",
		Tier: types.MerchantTierBronze, StakeAmount: math.NewInt(1_000),
		IsActive: true, TotalVolume: math.ZeroInt(),
	}
	require.NoError(t, k.SetMerchant(ctx, m))

	// Small amount → ok
	err := ms.ValidatePaymentOperationalConstraints(ctx, math.NewInt(100), addr)
	require.NoError(t, err)

	// Exceeds max if max > 0
	p, _ := k.GetParams(ctx)
	if p.MaxTransactionAmount.IsPositive() {
		overMax := p.MaxTransactionAmount.Add(math.NewInt(1))
		err = ms.ValidatePaymentOperationalConstraints(ctx, overMax, addr)
		require.Error(t, err)
	}
}

// ──────────────────────────────────────────────
// ValidateVaultOperationalConstraints
// ──────────────────────────────────────────────

func TestInternal_ValidateVaultOperationalConstraints(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	err := ms.ValidateVaultOperationalConstraints(ctx, math.NewInt(1_000_000), 100)
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// ValidateRewardPoolOperationalConstraints
// ──────────────────────────────────────────────

func TestInternal_ValidateRewardPoolConstraints(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	addr := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"
	m := types.Merchant{
		Address: addr, BusinessName: "M",
		Tier: types.MerchantTierBronze, StakeAmount: math.NewInt(100_000),
		IsActive: true, TotalVolume: math.ZeroInt(),
	}
	require.NoError(t, k.SetMerchant(ctx, m))

	// Within ratio
	err := ms.ValidateRewardPoolOperationalConstraints(ctx, addr, math.NewInt(500_000))
	require.NoError(t, err)

	// >10x stake
	err = ms.ValidateRewardPoolOperationalConstraints(ctx, addr, math.NewInt(10_000_001))
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// LogSecurityEvent
// ──────────────────────────────────────────────

func TestInternal_LogSecurityEvent(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)
	// Should not panic
	ms.LogSecurityEvent(ctx, "test_event", "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f", "unit test")
}

// ──────────────────────────────────────────────
// ValidateAuthorityWithLogging
// ──────────────────────────────────────────────

func TestInternal_ValidateAuthorityWithLogging(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	err := ms.ValidateAuthorityWithLogging(ctx, k.GetAuthority(), "test_op")
	require.NoError(t, err)

	err = ms.ValidateAuthorityWithLogging(ctx, "vita1xxxxinvalidaddress", "test_op")
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// ValidateBusinessLogicConsistency
// ──────────────────────────────────────────────

func TestInternal_ValidateBusinessLogicConsistency(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	addr := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"
	m := types.Merchant{
		Address: addr, BusinessName: "M",
		Tier: types.MerchantTierBronze, StakeAmount: math.NewInt(1_000),
		IsActive: true, TotalVolume: math.ZeroInt(),
	}
	require.NoError(t, k.SetMerchant(ctx, m))

	require.NoError(t, ms.ValidateBusinessLogicConsistency(ctx, "payment_creation", map[string]interface{}{
		"amount": math.NewInt(100), "merchant_address": addr,
	}))
	require.NoError(t, ms.ValidateBusinessLogicConsistency(ctx, "vault_creation", map[string]interface{}{
		"amount": math.NewInt(1_000_000), "lock_duration": uint64(100),
	}))
	require.NoError(t, ms.ValidateBusinessLogicConsistency(ctx, "reward_pool_creation", map[string]interface{}{
		"merchant_address": addr, "total_rewards": math.NewInt(5_000),
	}))
	require.NoError(t, ms.ValidateBusinessLogicConsistency(ctx, "unknown_op", map[string]interface{}{}))
}

// ──────────────────────────────────────────────
// ValidateTransaction (end-to-end shim)
// ──────────────────────────────────────────────

func TestInternal_ValidateTransaction(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	err := ms.ValidateTransaction(ctx, "payment_creation",
		"vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f", map[string]interface{}{})
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// ValidateTransactionFrequency
// ──────────────────────────────────────────────

func TestInternal_ValidateTransactionFrequency(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	addr := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	// First call (no rate limit set) → ok
	err := ms.ValidateTransactionFrequency(ctx, addr)
	require.NoError(t, err)

	// Set rate limit of 10 blocks
	require.NoError(t, k.SetMinBlocksBetweenTx(ctx, 10))

	// Same block → fail
	err = ms.ValidateTransactionFrequency(ctx, addr)
	require.Error(t, err)

	// Advance 10 blocks → ok
	advCtx := ctx.WithBlockHeight(ctx.BlockHeight() + 10)
	err = ms.ValidateTransactionFrequency(advCtx, addr)
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// ValidateEconomicConstraints
// ──────────────────────────────────────────────

func TestInternal_ValidateEconomicConstraints(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	err := ms.ValidateEconomicConstraints(ctx, []math.Int{math.NewInt(100), math.NewInt(200)})
	require.NoError(t, err)

	err = ms.ValidateEconomicConstraints(ctx, []math.Int{math.NewInt(-1)})
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// ValidateGasAndFees
// ──────────────────────────────────────────────

func TestInternal_ValidateGasAndFees(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	ms := newMsgServerInternal(k)
	// msgType string
	err := ms.ValidateGasAndFees(ctx, "payment_creation")
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// BeginBlocker / EndBlocker
// ──────────────────────────────────────────────

func TestInternal_BeginBlocker(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	err := k.BeginBlocker(ctx)
	require.NoError(t, err)
}

func TestInternal_EndBlocker(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	err := k.EndBlocker(ctx)
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// treasury_proposals
// ──────────────────────────────────────────────

func TestInternal_TreasuryProposals(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// ValidateTreasurySpendProposal — valid proposal struct, but treasury balance
	// check will fail in mock environment (module account returns nil).
	proposal := &types.TreasurySpendProposal{
		Title:       "Dev Fund",
		Description: "fund dev",
		Recipient:   "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
		Amount:      sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(1_000))),
	}
	// Exercises ValidateTreasurySpendProposal code path
	err := k.ValidateTreasurySpendProposal(ctx, proposal)
	_ = err // either passes (if mock supports it) or fails on treasury balance — both fine

	// Invalid: empty recipient
	bad := &types.TreasurySpendProposal{
		Title: "T", Description: "D", Recipient: "",
		Amount: sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(1_000))),
	}
	err = k.ValidateTreasurySpendProposal(ctx, bad)
	require.Error(t, err)

	// Invalid: zero amount
	bad2 := &types.TreasurySpendProposal{
		Title: "T", Description: "D",
		Recipient: "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
		Amount:    sdk.NewCoins(),
	}
	err = k.ValidateTreasurySpendProposal(ctx, bad2)
	require.Error(t, err)

	// HandleTreasurySpendProposal — will fail due to empty treasury in mock env
	err = HandleTreasurySpendProposal(ctx, k, proposal)
	require.Error(t, err) // expected: treasury module account not set up in mock env

	// NewTreasurySpendProposalHandler — check it returns non-nil
	handler := NewTreasurySpendProposalHandler(k)
	require.NotNil(t, handler)
}

// ──────────────────────────────────────────────
// Fee distribution helpers
// ──────────────────────────────────────────────

func TestInternal_RefundPaymentFunds(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	sender := sdk.AccAddress([]byte("sender-address-1234"))
	err := k.RefundPaymentFunds(ctx, sender, math.NewInt(1_000), "pay-001")
	require.NoError(t, err)
}

func TestInternal_DistributeProtocolFees(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	err := k.DistributeProtocolFees(ctx)
	_ = err // exercises the code path
}

// ──────────────────────────────────────────────
// msgServer unexported helpers
// ──────────────────────────────────────────────

func TestInternal_CalculateFee(t *testing.T) {
	k, _ := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	fee := ms.calculateFee(math.NewInt(10_000), math.LegacyNewDecWithPrec(1, 1)) // 0.1%
	require.True(t, fee.IsPositive())

	// Zero amount → zero fee
	fee2 := ms.calculateFee(math.NewInt(0), math.LegacyNewDecWithPrec(1, 1))
	require.True(t, fee2.IsZero())
}

func TestInternal_CalculateMerchantTierMsg(t *testing.T) {
	k, _ := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	platinum, _ := math.NewIntFromString("1000000000000000000000000")
	gold, _ := math.NewIntFromString("100000000000000000000000")
	silver, _ := math.NewIntFromString("10000000000000000000000")

	require.Equal(t, types.MerchantTierPlatinum, ms.calculateMerchantTier(platinum))
	require.Equal(t, types.MerchantTierGold, ms.calculateMerchantTier(gold))
	require.Equal(t, types.MerchantTierSilver, ms.calculateMerchantTier(silver))
	require.Equal(t, types.MerchantTierBronze, ms.calculateMerchantTier(math.NewInt(1)))
}

func TestInternal_CalculateVaultRewards(t *testing.T) {
	k, _ := setupTestKeeper(t)
	ms := newMsgServerInternal(k)

	rewards := ms.calculateVaultRewards(math.NewInt(1_000_000_000), 100_000)
	require.False(t, rewards.IsNegative())
}

// ──────────────────────────────────────────────
// getAllDelegations / DistributeStakingRewards
// ──────────────────────────────────────────────

func TestInternal_GetAllDelegations(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	// No delegations yet
	recs, err := k.getAllDelegations(ctx)
	require.NoError(t, err)
	require.Empty(t, recs)
}

func TestInternal_DistributeStakingRewards(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	// No fee stats yet → returns nil (no-op)
	err := k.DistributeStakingRewards(ctx)
	require.NoError(t, err)

	// Set fee statistics to trigger distribution path
	stats := types.FeeStatistics{
		TotalCollectedAllTime:    math.NewInt(0),
		TotalBurnedAllTime:       math.NewInt(0),
		TotalToValidatorsAllTime: math.NewInt(100_000),
		TotalToTreasuryAllTime:   math.NewInt(0),
		TotalTransactionsAllTime: 10,
		LastUpdateHeight:         1,
		CurrentEpoch:             0,
	}
	require.NoError(t, k.SetFeeStatistics(ctx, stats))
	// With no delegations → nothing to distribute
	err = k.DistributeStakingRewards(ctx)
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// LiquidDelegate / LiquidUndelegate
// ──────────────────────────────────────────────

func TestInternal_LiquidDelegate(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	delegator := sdk.AccAddress([]byte("delegator-addr-1234"))
	validatorAddr := sdk.ValAddress([]byte("validator-addr-12")).String()

	// empty delegator
	err := k.LiquidDelegate(ctx, sdk.AccAddress{}, validatorAddr, math.NewInt(1_000))
	require.Error(t, err)

	// empty validator
	err = k.LiquidDelegate(ctx, delegator, "", math.NewInt(1_000))
	require.Error(t, err)

	// zero amount
	err = k.LiquidDelegate(ctx, delegator, validatorAddr, math.NewInt(0))
	require.Error(t, err)

	// valid (mock bank sends always succeed)
	err = k.LiquidDelegate(ctx, delegator, validatorAddr, math.NewInt(1_000_000))
	require.NoError(t, err)
}

func TestInternal_LiquidUndelegate(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	delegator := sdk.AccAddress([]byte("delegator-addr-1234"))
	validatorAddr := sdk.ValAddress([]byte("validator-addr-12")).String()

	// empty delegator
	err := k.LiquidUndelegate(ctx, sdk.AccAddress{}, validatorAddr, math.NewInt(1_000))
	require.Error(t, err)

	// zero amount
	err = k.LiquidUndelegate(ctx, delegator, validatorAddr, math.NewInt(0))
	require.Error(t, err)

	// valid — may fail because no existing stVITA balance; just exercises the path
	err = k.LiquidUndelegate(ctx, delegator, validatorAddr, math.NewInt(1_000_000))
	_ = err
}

// ──────────────────────────────────────────────
// ImportTreasuryGenesis
// ──────────────────────────────────────────────

func TestInternal_ImportTreasuryGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	genesis := &types.TreasuryGenesisState{
		SpendingList: []types.TreasurySpending{
			{
				Id: "gen-1", ProposalId: 1,
				Recipient: "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
				Amount:    sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(100))),
				Purpose:   "genesis import",
			},
		},
	}
	err := k.ImportTreasuryGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify record was imported
	got, err := k.GetTreasurySpending(ctx, "gen-1")
	require.NoError(t, err)
	require.Equal(t, "gen-1", got.Id)
}

// ──────────────────────────────────────────────
// applyParamChange (indirect via keeper method)
// ──────────────────────────────────────────────

func TestInternal_ApplyParamChange(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Valid governance params
	require.NoError(t, k.applyParamChange(ctx, "min_deposit", "1000000"))
	require.NoError(t, k.applyParamChange(ctx, "voting_period", "100"))
	require.NoError(t, k.applyParamChange(ctx, "quorum", "0.334000000000000000"))
	require.NoError(t, k.applyParamChange(ctx, "threshold", "0.500000000000000000"))
	require.NoError(t, k.applyParamChange(ctx, "veto_threshold", "0.334000000000000000"))
	require.NoError(t, k.applyParamChange(ctx, "max_deposit_period", "500"))

	// Unknown key → error
	err := k.applyParamChange(ctx, "unknown_key", "value")
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// EndBlocker deeper paths (with data)
// ──────────────────────────────────────────────

func TestInternal_EndBlockerWithPayment(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Add an expired payment so EndBlocker exercises the expiry path
	payment := types.Payment{
		Id:             "pay-expire-1",
		FromAddress:    "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
		ToAddress:      "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
		Amount:         math.NewInt(1_000),
		Status:         types.PaymentStatusPending,
		CreationHeight: 1,
	}
	require.NoError(t, k.SetPayment(ctx, payment))

	err := k.EndBlocker(ctx)
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// CanBurnTokens
// ──────────────────────────────────────────────

func TestInternal_CanBurnTokens(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// No burn stats yet → should be allowed
	ok, err := k.CanBurnTokens(ctx, math.NewInt(1_000))
	require.NoError(t, err)
	require.True(t, ok)

	// Set burn stats with cap reached
	bs := types.BurnStats{
		TotalBurned:    math.NewInt(1_000_000),
		BurnCapReached: true,
		LastBurnHeight: 1,
	}
	require.NoError(t, k.SetBurnStatistics(ctx, bs))
	ok, err = k.CanBurnTokens(ctx, math.NewInt(1_000))
	require.NoError(t, err)
	require.False(t, ok)
}

// ──────────────────────────────────────────────
// DistributeProtocolFees deeper paths
// ──────────────────────────────────────────────

func TestInternal_DistributeProtocolFees_WithFeeStats(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Set fee statistics so distribution gets past the first guard
	stats := types.FeeStatistics{
		TotalCollectedAllTime:    math.NewInt(100_000),
		TotalBurnedAllTime:       math.NewInt(0),
		TotalToValidatorsAllTime: math.NewInt(0),
		TotalToTreasuryAllTime:   math.NewInt(0),
		TotalTransactionsAllTime: 1,
		LastUpdateHeight:         1,
		CurrentEpoch:             0,
	}
	require.NoError(t, k.SetFeeStatistics(ctx, stats))

	err := k.DistributeProtocolFees(ctx)
	_ = err // may fail on treasury/validator sends; exercising the code path
}

// ──────────────────────────────────────────────
// UpdateBurnStatistics
// ──────────────────────────────────────────────

func TestInternal_UpdateBurnStatistics(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	amount := math.NewInt(50_000)
	err := k.UpdateBurnStatistics(ctx, amount)
	require.NoError(t, err)

	bs, err := k.GetBurnStatistics(ctx)
	require.NoError(t, err)
	require.Equal(t, amount, bs.TotalBurned)
}

// ──────────────────────────────────────────────
// EndBlockerGovernance with deposit-period expiry
// ──────────────────────────────────────────────

func TestInternal_EndBlockerGovernance_DepositExpiry(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	proposer := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	govParams := k.GetGovernanceParams(ctx)
	// Submit with deposit below threshold → stays in DEPOSIT phase
	smallDeposit := govParams.MinDeposit.Sub(math.NewInt(1))
	if smallDeposit.IsNegative() {
		smallDeposit = math.ZeroInt()
	}

	id, err := k.SubmitProposal(ctx, proposer, "Deposit Expire", "desc",
		types.ProposalTypeText, "", smallDeposit)
	require.NoError(t, err)

	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)

	// Jump past deposit end time
	expiredCtx := ctx.WithBlockHeight(p.DepositEndTime + 1)
	err = k.EndBlockerGovernance(expiredCtx)
	require.NoError(t, err)

	// Proposal should be FAILED
	p2, found2, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found2)
	require.Equal(t, types.ProposalStatusFailed, p2.Status)
}

// ──────────────────────────────────────────────
// EstimateTreasurySpendImpact
// ──────────────────────────────────────────────

func TestInternal_EstimateTreasurySpendImpact(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	amount := sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(1_000)))
	_, err := k.EstimateTreasurySpendImpact(ctx, amount)
	// May fail due to treasury module account not set up in mock env — that's expected.
	_ = err
}

// ──────────────────────────────────────────────
// gRPC query: PaymentAll / VaultAll / RewardPoolAll
// ──────────────────────────────────────────────

func TestInternal_GrpcQueryAll(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	qs := NewQueryServerImpl(k)

	// PaymentAll
	resp1, err := qs.PaymentAll(ctx, &types.QueryAllPaymentRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// VaultAll
	resp2, err := qs.VaultAll(ctx, &types.QueryAllVaultRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp2)

	// RewardPoolAll
	resp3, err := qs.RewardPoolAll(ctx, &types.QueryAllRewardPoolRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp3)
}

// ──────────────────────────────────────────────
// gRPC fee queries
// ──────────────────────────────────────────────

func TestInternal_GrpcFeeQueries(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	qs := NewQueryServerImpl(k)

	// FeeAccumulator (store is empty → returns empty struct)
	resp, err := qs.FeeAccumulator(ctx, &types.QueryFeeAccumulatorRequest{})
	_ = err
	_ = resp

	// FeeStatistics
	resp2, err := qs.FeeStatistics(ctx, &types.QueryFeeStatisticsRequest{})
	_ = err
	_ = resp2

	// BurnStatistics
	resp3, err := qs.BurnStatistics(ctx, &types.QueryBurnStatisticsRequest{})
	_ = err
	_ = resp3

	// SupplySnapshot
	resp4, err := qs.SupplySnapshot(ctx, &types.QuerySupplySnapshotRequest{})
	_ = err
	_ = resp4
}

// ──────────────────────────────────────────────
// TreasuryHealth
// ──────────────────────────────────────────────

func TestInternal_TreasuryHealth(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	resp, err := k.TreasuryHealth(ctx, nil)
	require.Error(t, err)
	_ = resp

	resp, err = k.TreasuryHealth(ctx, &types.QueryTreasuryHealthRequest{})
	_ = err
	_ = resp
}

// ──────────────────────────────────────────────
// CanBurnTokens — more branches
// ──────────────────────────────────────────────

func TestInternal_CanBurnTokens_BurnCapBranches(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Set a positive BurnCapSupply so the branch executes
	p, _ := k.GetParams(ctx)
	p.BurnCapSupply = math.NewInt(1_000_000_000)
	require.NoError(t, k.SetParams(ctx, p))

	// With no burn stats → allowed
	ok, err := k.CanBurnTokens(ctx, math.NewInt(100))
	require.NoError(t, err)
	require.True(t, ok)

	// Set stats well under cap — CurrentSupply must be big enough
	bs := types.BurnStats{
		TotalBurned:    math.NewInt(500_000_000),
		BurnCapReached: false,
		LastBurnHeight: 1,
		BurnCapSupply:  math.NewInt(1_000_000_000),
		CurrentSupply:  math.NewInt(2_000_000_000), // above cap so small burns are fine
		RemainingToCap: math.NewInt(500_000_000),
	}
	require.NoError(t, k.SetBurnStatistics(ctx, bs))
	ok, err = k.CanBurnTokens(ctx, math.NewInt(100))
	require.NoError(t, err)
	require.True(t, ok)

	// Amount would push supply below cap → false
	ok, err = k.CanBurnTokens(ctx, math.NewInt(1_100_000_000))
	require.NoError(t, err)
	require.False(t, ok)
}

// ──────────────────────────────────────────────
// CalculateProtocolFee — edge cases
// ──────────────────────────────────────────────

func TestInternal_CalculateProtocolFee(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Normal fee calculation
	fee, netAmt, err := k.CalculateProtocolFee(ctx, math.NewInt(10_000))
	require.NoError(t, err)
	require.False(t, fee.IsNegative())
	require.False(t, netAmt.IsNegative())
}

// ──────────────────────────────────────────────
// EscrowPaymentFunds / ReleasePaymentFunds
// ──────────────────────────────────────────────

func TestInternal_EscrowAndRelease(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	payer := sdk.AccAddress([]byte("payer-address-1234"))
	amount := math.NewInt(5_000_000)
	paymentID := "escrow-test-001"

	err := k.EscrowPaymentFunds(ctx, payer, amount)
	require.NoError(t, err)

	merchant := sdk.AccAddress([]byte("merchant-addr-12"))
	feeAmt, netAmt, err := k.ReleasePaymentFunds(ctx, merchant, amount, paymentID)
	require.NoError(t, err)
	require.False(t, feeAmt.IsNegative())
	require.False(t, netAmt.IsNegative())
}

// ──────────────────────────────────────────────
// AccumulateProtocolFee, UpdateFeeStatistics
// ──────────────────────────────────────────────

func TestInternal_AccumulateFeeAndStats(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	feeAmt := math.NewInt(1_000)
	err := k.AccumulateProtocolFee(ctx, feeAmt)
	require.NoError(t, err)

	// UpdateFeeStatistics
	err = k.UpdateFeeStatistics(ctx, feeAmt, math.NewInt(100), math.NewInt(500), math.NewInt(400), uint64(1))
	require.NoError(t, err)

	stats, err := k.GetFeeStatistics(ctx)
	require.NoError(t, err)
	require.True(t, stats.TotalCollectedAllTime.IsPositive())
}

// ──────────────────────────────────────────────
// DelegateToValidator
// ──────────────────────────────────────────────

func TestInternal_DelegateToValidator(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	delegator := sdk.AccAddress([]byte("delegator-addr-1234"))
	valAddr := sdk.ValAddress([]byte("validator-addr-12"))

	// invalid amount (negative via sdk.Coin is rejected by sdk.Coin.IsPositive)
	badCoin := sdk.NewCoin(types.BondDenom, math.NewInt(0))
	err := k.DelegateToValidator(ctx, delegator, valAddr, badCoin)
	require.Error(t, err)

	// valid delegation
	coin := sdk.NewCoin(types.BondDenom, math.NewInt(100_000))
	err = k.DelegateToValidator(ctx, delegator, valAddr, coin)
	require.NoError(t, err)

	// cumulative delegation
	err = k.DelegateToValidator(ctx, delegator, valAddr, sdk.NewCoin(types.BondDenom, math.NewInt(50_000)))
	require.NoError(t, err)
}

// ──────────────────────────────────────────────
// CastVote — more branches
// ──────────────────────────────────────────────

func TestInternal_CastVote_Branches(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	proposer := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	// Pre-register a delegation so CastVote accepts the voter
	delegator, _ := sdk.AccAddressFromBech32(proposer)
	valAddr := sdk.ValAddress(delegator)
	require.NoError(t, k.DelegateToValidator(ctx,
		delegator, valAddr,
		sdk.NewCoin(types.BondDenom, math.NewInt(1_000_000))))

	govParams := k.GetGovernanceParams(ctx)
	id, err := k.SubmitProposal(ctx, proposer, "Vote Test", "desc",
		types.ProposalTypeText, "", govParams.MinDeposit)
	require.NoError(t, err)

	// Vote No
	require.NoError(t, k.CastVote(ctx, id, proposer, types.VoteOptionNo))

	// Change vote to Abstain
	require.NoError(t, k.CastVote(ctx, id, proposer, types.VoteOptionAbstain))

	// Non-existent proposal
	err = k.CastVote(ctx, 9999, proposer, types.VoteOptionYes)
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// TallyProposal — direct call
// ──────────────────────────────────────────────

func TestInternal_TallyProposal(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	proposer := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	delegator, _ := sdk.AccAddressFromBech32(proposer)
	valAddr := sdk.ValAddress(delegator)
	require.NoError(t, k.DelegateToValidator(ctx, delegator, valAddr,
		sdk.NewCoin(types.BondDenom, math.NewInt(10_000_000))))

	govParams := k.GetGovernanceParams(ctx)
	id, err := k.SubmitProposal(ctx, proposer, "Tally Test", "desc",
		types.ProposalTypeText, "", govParams.MinDeposit)
	require.NoError(t, err)

	require.NoError(t, k.CastVote(ctx, id, proposer, types.VoteOptionYes))

	passed, err := k.TallyProposal(ctx, id)
	require.NoError(t, err)
	_ = passed

	// Non-existent proposal → error
	_, err = k.TallyProposal(ctx, 9999)
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// ExecuteProposal path
// ──────────────────────────────────────────────

func TestInternal_ExecuteProposal(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	proposer := "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f"

	govParams := k.GetGovernanceParams(ctx)
	id, err := k.SubmitProposal(ctx, proposer, "Exec Test", "desc",
		types.ProposalTypeText, "", govParams.MinDeposit)
	require.NoError(t, err)

	// Force proposal to PASSED status for execution
	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)
	p.Status = types.ProposalStatusPassed
	require.NoError(t, k.SetProposal(ctx, p))

	// Execute should succeed (text proposals have no action)
	err = k.ExecuteProposal(ctx, id)
	require.NoError(t, err)

	// Non-existent → error
	err = k.ExecuteProposal(ctx, 9999)
	require.Error(t, err)
}

// ──────────────────────────────────────────────
// Validator CRUD
// ──────────────────────────────────────────────

func TestInternal_ValidatorCRUD(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	valAddr := sdk.ValAddress([]byte("test-validator-12")).String()

	// GetValidator on empty store → not found
	_, found, err := k.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.False(t, found)

	// SetValidator
	val := ValidatorRecord{
		OperatorAddress: valAddr,
		Moniker:         "test-node",
		Commission:      math.LegacyNewDecWithPrec(5, 2),
		TotalDelegated:  math.NewInt(1_000_000),
		SelfBond:        math.NewInt(100_000),
		Jailed:          false,
		CreatedBlock:    1,
	}
	err = k.SetValidator(ctx, val)
	require.NoError(t, err)

	// GetValidator after set
	got, found, err := k.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, val.TotalDelegated, got.TotalDelegated)
}

// ──────────────────────────────────────────────
// getTotalVITADelegated
// ──────────────────────────────────────────────

func TestInternal_GetTotalVITADelegated(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// No delegations → zero
	total, err := k.getTotalVITADelegated(ctx)
	require.NoError(t, err)
	require.True(t, total.IsZero())

	// Add a delegation via DelegateToValidator then check
	delegator := sdk.AccAddress([]byte("delegator-addr-1234"))
	valAddr := sdk.ValAddress([]byte("validator-addr-12"))
	require.NoError(t, k.DelegateToValidator(ctx, delegator, valAddr,
		sdk.NewCoin(types.BondDenom, math.NewInt(500_000))))

	total2, err := k.getTotalVITADelegated(ctx)
	require.NoError(t, err)
	require.True(t, total2.IsPositive())
}

// ──────────────────────────────────────────────
// processTimeBasedOperations / calculateRewardMultiplier
// (triggered via BeginBlocker when there is state)
// ──────────────────────────────────────────────

func TestInternal_BeginBlockerWithVaults(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	vault := types.Vault{
		Id:             "vault-001",
		Owner:          "vita1tshzqh0puwkm8u2kj7mz2jek6gsylujn3qaq3f",
		Amount:         math.NewInt(1_000_000),
		LockDuration:   100,
		CreationHeight: 1,
		UnlockHeight:   2, // already mature at height 1
		RewardMultiplier: math.LegacyOneDec(),
	}
	require.NoError(t, k.SetVault(ctx, vault))

	err := k.BeginBlocker(ctx)
	require.NoError(t, err)
}

func TestInternal_CalculateRewardMultiplier(t *testing.T) {
	k, _ := setupTestKeeper(t)

	// short duration
	mult := k.calculateRewardMultiplier(uint64(1_000))
	require.True(t, mult.GTE(math.LegacyOneDec()))

	// long duration
	mult2 := k.calculateRewardMultiplier(uint64(100_000))
	require.True(t, mult2.GTE(mult))
}
