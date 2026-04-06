package keeper_test

import (
	"encoding/json"
	"testing"
	"time"

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

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newGovTestContext creates a fresh keeper + context for governance unit tests.
// block height is set to 100 so VotingEndTime can be set far in the future.
func newGovTestContext(t *testing.T) (sdk.Context, keeper.Keeper, *MockBankKeeper) {
	t.Helper()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")
	config.SetBech32PrefixForValidator("vitavaloper", "vitavaloperpub")
	config.SetBech32PrefixForConsensusNode("vitavalcons", "vitavalconspub")

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 100, Time: time.Now()}, false, log.NewNopLogger())

	bankKeeper := NewMockBankKeeper()
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		bankKeeper,
		&MockAccountKeeper{},
	)

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	return ctx, k, bankKeeper
}

// govMinDeposit returns a deposit amount that meets MinDeposit.
func govMinDeposit() math.Int {
	d, _ := math.NewIntFromString("10000000000000000000000") // 10_000 * 10^18
	return d
}

// govBelowDeposit returns a deposit amount below MinDeposit.
func govBelowDeposit() math.Int {
	d, _ := math.NewIntFromString("1000000000000000000000") // 1_000 * 10^18
	return d
}

// govProposerAddr returns a deterministic vita-prefixed proposer address.
func govProposerAddr(t *testing.T) (sdk.AccAddress, string) {
	t.Helper()
	addr := sdk.AccAddress([]byte("govproposer123456"))
	return addr, addr.String()
}

// govVoterAddr returns a deterministic vita-prefixed voter address.
func govVoterAddr(t *testing.T) (sdk.AccAddress, string) {
	t.Helper()
	addr := sdk.AccAddress([]byte("govvoteraddr12345"))
	return addr, addr.String()
}

// govRecipientAddr returns a deterministic vita-prefixed recipient address.
func govRecipientAddr(t *testing.T) (sdk.AccAddress, string) {
	t.Helper()
	addr := sdk.AccAddress([]byte("govrecipient12345"))
	return addr, addr.String()
}

// setupValidatorAndDelegation registers a validator and a delegation so that
// getVoterStake returns a non-zero weight for voterAddrStr.
// RegisterValidator accepts AccAddress bech32 (vita prefix), not ValAddress.
func setupValidatorAndDelegation(
	t *testing.T,
	ctx sdk.Context,
	k keeper.Keeper,
	voterAddrStr string,
	delegateAmt math.Int,
) {
	t.Helper()
	// Use AccAddress for the validator operator (RegisterValidator calls AccAddressFromBech32)
	valAccAddr := sdk.AccAddress([]byte("govvalidator12345"))
	valAddrStr := valAccAddr.String() // vita-prefixed bech32

	// Register validator with self-bond meeting minimum
	selfBond := types.DefaultStakingParams().MinValidatorBond
	err := k.RegisterValidator(ctx, valAddrStr, "gov-test-val",
		math.LegacyMustNewDecFromStr("0.05"), selfBond)
	require.NoError(t, err)

	// Delegate from voter so getVoterStake has data.
	// DelegateToValidator takes ValAddress; we derive it from the same bytes.
	voterAcc, err := sdk.AccAddressFromBech32(voterAddrStr)
	require.NoError(t, err)
	valAddr := sdk.ValAddress(valAccAddr)
	coin := sdk.NewCoin(types.BondDenom, delegateAmt)
	err = k.DelegateToValidator(ctx, voterAcc, valAddr, coin)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// TestSubmitProposal
// ---------------------------------------------------------------------------

func TestSubmitProposal(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)

	deposit := govMinDeposit()

	id, err := k.SubmitProposal(ctx,
		proposerStr,
		"Test Proposal",
		"A text proposal for testing",
		types.ProposalTypeText,
		"",
		deposit,
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, id, "first proposal should have id=1")

	// Verify stored proposal
	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found, "proposal should be findable after submission")

	// Should be in Voting status because deposit >= MinDeposit
	require.Equal(t, types.ProposalStatusVoting, p.Status,
		"proposal with sufficient deposit should advance to Voting")
	require.Equal(t, "Test Proposal", p.Title)
	require.Equal(t, proposerStr, p.Proposer)
	require.True(t, p.TotalDeposit.Equal(deposit))
}

// ---------------------------------------------------------------------------
// TestSubmitProposal_InsufficientDeposit
// ---------------------------------------------------------------------------

func TestSubmitProposal_InsufficientDeposit(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)

	deposit := govBelowDeposit()

	id, err := k.SubmitProposal(ctx,
		proposerStr,
		"Underfunded Proposal",
		"Deposit is below MinDeposit",
		types.ProposalTypeText,
		"",
		deposit,
	)
	require.NoError(t, err)

	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)

	// Should remain in Deposit status
	require.Equal(t, types.ProposalStatusDeposit, p.Status,
		"proposal with insufficient deposit should stay in Deposit status")
	require.EqualValues(t, 0, p.VotingEndTime,
		"VotingEndTime should not be set for deposit-only proposals")
}

// ---------------------------------------------------------------------------
// TestCastVote
// ---------------------------------------------------------------------------

func TestCastVote(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)
	_, voterStr := govVoterAddr(t)

	// Give the voter a delegation so getVoterStake > 0
	stakeAmt := math.NewInt(1_000_000).Mul(math.NewInt(1_000_000_000_000_000_000)) // 1M VITA
	setupValidatorAndDelegation(t, ctx, k, voterStr, stakeAmt)

	deposit := govMinDeposit()
	id, err := k.SubmitProposal(ctx, proposerStr, "Vote Test", "desc",
		types.ProposalTypeText, "", deposit)
	require.NoError(t, err)

	// Cast a Yes vote
	err = k.CastVote(ctx, id, voterStr, types.VoteOptionYes)
	require.NoError(t, err)

	// Verify vote stored
	v, found, err := k.GetVote(ctx, id, voterStr)
	require.NoError(t, err)
	require.True(t, found, "vote should be stored after CastVote")
	require.Equal(t, int32(types.VoteOptionYes), v.Option)
	require.Equal(t, voterStr, v.Voter)
	require.True(t, v.Weight.IsPositive(), "vote weight should be positive")
}

// ---------------------------------------------------------------------------
// TestCastVote_ProposalNotVoting
// ---------------------------------------------------------------------------

func TestCastVote_ProposalNotVoting(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)
	_, voterStr := govVoterAddr(t)

	// Submit with insufficient deposit → stays in Deposit status
	deposit := govBelowDeposit()
	id, err := k.SubmitProposal(ctx, proposerStr, "Not Voting", "desc",
		types.ProposalTypeText, "", deposit)
	require.NoError(t, err)

	// CastVote should fail because proposal is not in Voting status
	err = k.CastVote(ctx, id, voterStr, types.VoteOptionYes)
	require.Error(t, err, "voting on a deposit-phase proposal should fail")
	require.Contains(t, err.Error(), "not in voting period")
}

// ---------------------------------------------------------------------------
// TestTallyProposal_Passes
// ---------------------------------------------------------------------------

func TestTallyProposal_Passes(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)
	_, voterStr := govVoterAddr(t)

	// Large stake so quorum is met (must exceed 33.4% of totalStaked = selfBond+delegateAmt)
	// selfBond = 10_000 VITA; delegateAmt = 1_000_000 VITA → delegateAmt/totalStaked ≈ 99% > 33.4%
	stakeAmt := math.NewInt(1_000_000).Mul(math.NewInt(1_000_000_000_000_000_000)) // 1M VITA
	setupValidatorAndDelegation(t, ctx, k, voterStr, stakeAmt)

	deposit := govMinDeposit()
	id, err := k.SubmitProposal(ctx, proposerStr, "Passing Proposal", "desc",
		types.ProposalTypeText, "", deposit)
	require.NoError(t, err)

	// Cast Yes vote
	err = k.CastVote(ctx, id, voterStr, types.VoteOptionYes)
	require.NoError(t, err)

	// Advance block past voting end time to allow tally
	p, _, _ := k.GetProposal(ctx, id)
	ctx = ctx.WithBlockHeight(p.VotingEndTime + 1)

	passed, err := k.TallyProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, passed, "proposal with all Yes votes should pass")

	// Verify status
	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, types.ProposalStatusPassed, p.Status)
}

// ---------------------------------------------------------------------------
// TestTallyProposal_Vetoed
// ---------------------------------------------------------------------------

func TestTallyProposal_Vetoed(t *testing.T) {
	ctx, k, _ := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)
	_, voterStr := govVoterAddr(t)

	stakeAmt := math.NewInt(1_000_000).Mul(math.NewInt(1_000_000_000_000_000_000)) // 1M VITA
	setupValidatorAndDelegation(t, ctx, k, voterStr, stakeAmt)

	deposit := govMinDeposit()
	id, err := k.SubmitProposal(ctx, proposerStr, "Veto Proposal", "desc",
		types.ProposalTypeText, "", deposit)
	require.NoError(t, err)

	// Cast NoWithVeto vote (>33.4% of total votes → veto)
	err = k.CastVote(ctx, id, voterStr, types.VoteOptionNoWithVeto)
	require.NoError(t, err)

	// Advance block past voting end time
	p, _, _ := k.GetProposal(ctx, id)
	ctx = ctx.WithBlockHeight(p.VotingEndTime + 1)

	passed, err := k.TallyProposal(ctx, id)
	require.NoError(t, err)
	require.False(t, passed, "all-veto proposal should not pass")

	// Verify rejected
	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, types.ProposalStatusRejected, p.Status)
}

// ---------------------------------------------------------------------------
// TestExecuteProposal_TreasurySpend
// ---------------------------------------------------------------------------

func TestExecuteProposal_TreasurySpend(t *testing.T) {
	ctx, k, bank := newGovTestContext(t)
	_, proposerStr := govProposerAddr(t)
	_, voterStr := govVoterAddr(t)
	_, recipientStr := govRecipientAddr(t)

	stakeAmt := math.NewInt(1_000_000).Mul(math.NewInt(1_000_000_000_000_000_000)) // 1M VITA
	setupValidatorAndDelegation(t, ctx, k, voterStr, stakeAmt)

	// Build treasury_spend content
	spendAmt := math.NewInt(500_000_000_000_000_000) // 0.5 VITA
	contentMap := map[string]string{
		"recipient": recipientStr,
		"amount":    spendAmt.String(),
		"denom":     types.BondDenom,
	}
	contentBytes, err := json.Marshal(contentMap)
	require.NoError(t, err)

	deposit := govMinDeposit()
	id, err := k.SubmitProposal(ctx, proposerStr, "Treasury Spend", "send funds",
		types.ProposalTypeTreasurySpend, string(contentBytes), deposit)
	require.NoError(t, err)

	// Cast Yes vote
	err = k.CastVote(ctx, id, voterStr, types.VoteOptionYes)
	require.NoError(t, err)

	// Advance past voting end
	p, _, _ := k.GetProposal(ctx, id)
	ctx = ctx.WithBlockHeight(p.VotingEndTime + 1)

	// Tally → should pass
	passed, err := k.TallyProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, passed)

	// Execute
	err = k.ExecuteProposal(ctx, id)
	require.NoError(t, err)

	// Verify recipient received funds via bank mock (SendCoinsFromModuleToAccount called)
	// MockBankKeeper.SendCoinsFromModuleToAccount always succeeds; verify proposal executed.
	p, found, err := k.GetProposal(ctx, id)
	require.NoError(t, err)
	require.True(t, found)
	// After successful execution the proposal stays Passed (ExecuteProposal only flips to Failed on error)
	require.NotEqual(t, types.ProposalStatusFailed, p.Status,
		"treasury_spend proposal should not be in Failed status after successful execution")

	// Verify that the deposit was returned to proposer by TallyProposal
	// (MockBankKeeper tracks SendCoinsFromAccountToModule but not sends out; just ensure no error path)
	_ = bank
}
