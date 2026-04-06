package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ───────────────────────────────────────────────────────────────────────────────
// Phase 5 — Governance: Proposal counter
// ───────────────────────────────────────────────────────────────────────────────

// GetProposalCounter returns the current proposal sequence counter.
// Returns 0 if no proposal has been submitted yet.
func (k Keeper) GetProposalCounter(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.KeyProposalCounter)
	if err != nil {
		return 0, fmt.Errorf("failed to get proposal counter: %w", err)
	}
	if bz == nil {
		return 0, nil
	}
	if len(bz) != 8 {
		return 0, fmt.Errorf("corrupt proposal counter: expected 8 bytes, got %d", len(bz))
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetProposalCounter persists the proposal sequence counter.
func (k Keeper) SetProposalCounter(ctx context.Context, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return store.Set(types.KeyProposalCounter, bz)
}

// nextProposalId increments the counter and returns the new proposal ID.
func (k Keeper) nextProposalId(ctx context.Context) (uint64, error) {
	current, err := k.GetProposalCounter(ctx)
	if err != nil {
		return 0, err
	}
	next := current + 1
	if err := k.SetProposalCounter(ctx, next); err != nil {
		return 0, err
	}
	return next, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// Proposal CRUD
// ───────────────────────────────────────────────────────────────────────────────

// SetProposal serialises a Proposal to JSON and writes it to the KV store.
func (k Keeper) SetProposal(ctx context.Context, p types.Proposal) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid proposal: %w", err)
	}
	bz, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal %d: %w", p.ProposalId, err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetProposalKey(p.ProposalId), bz)
}

// GetProposal retrieves a proposal by ID.
// Returns (proposal, true, nil) when found, (zero, false, nil) when not found.
func (k Keeper) GetProposal(ctx context.Context, proposalId uint64) (types.Proposal, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetProposalKey(proposalId))
	if err != nil {
		return types.Proposal{}, false, fmt.Errorf("failed to get proposal %d: %w", proposalId, err)
	}
	if bz == nil {
		return types.Proposal{}, false, nil
	}
	var p types.Proposal
	if err := json.Unmarshal(bz, &p); err != nil {
		return types.Proposal{}, false, fmt.Errorf("failed to unmarshal proposal %d: %w", proposalId, err)
	}
	return p, true, nil
}

// GetAllProposals iterates the proposal prefix and returns all proposals.
func (k Keeper) GetAllProposals(ctx context.Context) ([]types.Proposal, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.KeyPrefixProposal,
		storetypes.PrefixEndBytes(types.KeyPrefixProposal),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open proposal iterator: %w", err)
	}
	defer iter.Close()

	var proposals []types.Proposal
	for ; iter.Valid(); iter.Next() {
		var p types.Proposal
		if err := json.Unmarshal(iter.Value(), &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposal: %w", err)
		}
		proposals = append(proposals, p)
	}
	return proposals, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// Vote CRUD
// ───────────────────────────────────────────────────────────────────────────────

// SetVote serialises a Vote to JSON and writes it to the KV store.
func (k Keeper) SetVote(ctx context.Context, v types.Vote) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid vote: %w", err)
	}
	bz, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal vote for proposal %d voter %s: %w", v.ProposalId, v.Voter, err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetVoteKey(v.ProposalId, v.Voter), bz)
}

// GetVote retrieves a vote.
// Returns (vote, true, nil) when found, (zero, false, nil) when not found.
func (k Keeper) GetVote(ctx context.Context, proposalId uint64, voter string) (types.Vote, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetVoteKey(proposalId, voter))
	if err != nil {
		return types.Vote{}, false, fmt.Errorf("failed to get vote: %w", err)
	}
	if bz == nil {
		return types.Vote{}, false, nil
	}
	var v types.Vote
	if err := json.Unmarshal(bz, &v); err != nil {
		return types.Vote{}, false, fmt.Errorf("failed to unmarshal vote: %w", err)
	}
	return v, true, nil
}

// GetVotesByProposal returns all votes cast on a proposal.
func (k Keeper) GetVotesByProposal(ctx context.Context, proposalId uint64) ([]types.Vote, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetVotesByProposalPrefix(proposalId)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to open vote iterator for proposal %d: %w", proposalId, err)
	}
	defer iter.Close()

	var votes []types.Vote
	for ; iter.Valid(); iter.Next() {
		var v types.Vote
		if err := json.Unmarshal(iter.Value(), &v); err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote: %w", err)
		}
		votes = append(votes, v)
	}
	return votes, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// Governance Parameters
// ───────────────────────────────────────────────────────────────────────────────

// GetGovernanceParams returns governance parameters.
// Phase 5 uses hardcoded defaults; a future job will make these chain-updateable.
func (k Keeper) GetGovernanceParams(_ context.Context) types.GovernanceParams {
	return types.DefaultGovernanceParams()
}

// ───────────────────────────────────────────────────────────────────────────────
// SubmitProposal
// ───────────────────────────────────────────────────────────────────────────────

// SubmitProposal creates and stores a new governance proposal.
//
// Flow:
//  1. Validate inputs and deposit amount.
//  2. Lock deposit by transferring coins from proposer to the governance escrow
//     (vitacoin_treasury module account).
//  3. Assign a new proposal ID and set DepositEndTime = currentBlock + MaxDepositPeriod.
//  4. If initial deposit meets MinDeposit, auto-advance to VotingPeriod.
//  5. Emit EventTypeProposalSubmitted.
//  6. Return new proposalId.
func (k Keeper) SubmitProposal(
	ctx sdk.Context,
	proposer, title, description, proposalType, content string,
	deposit math.Int,
) (uint64, error) {
	// ── Input validation ──────────────────────────────────────────────────────
	if proposer == "" {
		return 0, fmt.Errorf("proposer address cannot be empty")
	}
	if title == "" {
		return 0, fmt.Errorf("proposal title cannot be empty")
	}
	if description == "" {
		return 0, fmt.Errorf("proposal description cannot be empty")
	}
	if proposalType != types.ProposalTypeText &&
		proposalType != types.ProposalTypeParamChange &&
		proposalType != types.ProposalTypeTreasurySpend {
		return 0, fmt.Errorf("unknown proposal type: %s", proposalType)
	}
	if deposit.IsNil() || deposit.IsNegative() {
		return 0, fmt.Errorf("deposit must be a non-negative value")
	}

	govParams := k.GetGovernanceParams(ctx)

	// ── Validate proposer address ─────────────────────────────────────────────
	proposerAddr, err := sdk.AccAddressFromBech32(proposer)
	if err != nil {
		return 0, fmt.Errorf("invalid proposer address: %w", err)
	}

	// ── Lock deposit ──────────────────────────────────────────────────────────
	if deposit.IsPositive() {
		depositCoins := sdk.NewCoins(sdk.NewCoin(types.BondDenom, deposit))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx,
			proposerAddr,
			types.TreasuryModuleName,
			depositCoins,
		); err != nil {
			return 0, fmt.Errorf("failed to lock deposit: %w", err)
		}
	}

	// ── Assign ID and build proposal ──────────────────────────────────────────
	proposalId, err := k.nextProposalId(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to assign proposal ID: %w", err)
	}

	currentBlock := ctx.BlockHeight()
	currentTime := ctx.BlockTime().Unix()

	proposal := types.Proposal{
		ProposalId:     proposalId,
		Title:          title,
		Description:    description,
		ProposalType:   proposalType,
		Status:         types.ProposalStatusDeposit,
		Proposer:       proposer,
		SubmitTime:     currentTime,
		DepositEndTime: currentBlock + govParams.MaxDepositPeriod,
		VotingEndTime:  0, // set when entering voting
		TotalDeposit:   deposit,
		YesVotes:       math.LegacyZeroDec(),
		NoVotes:        math.LegacyZeroDec(),
		AbstainVotes:   math.LegacyZeroDec(),
		VetoVotes:      math.LegacyZeroDec(),
		Content:        content,
	}

	// ── Auto-advance to voting if deposit threshold met ───────────────────────
	if deposit.GTE(govParams.MinDeposit) {
		proposal.Status = types.ProposalStatusVoting
		proposal.VotingEndTime = currentBlock + govParams.VotingPeriod
	}

	if err := k.SetProposal(ctx, proposal); err != nil {
		return 0, fmt.Errorf("failed to store proposal: %w", err)
	}

	// ── Emit event ────────────────────────────────────────────────────────────
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalSubmitted,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
			sdk.NewAttribute(types.AttributeKeyProposer, proposer),
			sdk.NewAttribute(types.AttributeKeyProposalType, proposalType),
			sdk.NewAttribute(types.AttributeKeyProposalStatus, proposal.StatusString()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	k.logger.Info("Governance proposal submitted",
		"proposal_id", proposalId,
		"proposer", proposer,
		"type", proposalType,
		"status", proposal.StatusString(),
		"deposit", deposit.String(),
	)

	return proposalId, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// AddDeposit
// ───────────────────────────────────────────────────────────────────────────────

// AddDeposit adds a deposit to an existing proposal in the DEPOSIT phase.
//
// If the new TotalDeposit meets or exceeds MinDeposit, the proposal is
// automatically advanced to the VOTING phase.
func (k Keeper) AddDeposit(
	ctx sdk.Context,
	proposalId uint64,
	depositor string,
	amount math.Int,
) error {
	if depositor == "" {
		return fmt.Errorf("depositor address cannot be empty")
	}
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("deposit amount must be positive")
	}

	// ── Fetch proposal ────────────────────────────────────────────────────────
	proposal, found, err := k.GetProposal(ctx, proposalId)
	if err != nil {
		return fmt.Errorf("failed to retrieve proposal %d: %w", proposalId, err)
	}
	if !found {
		return fmt.Errorf("proposal %d not found", proposalId)
	}
	if proposal.Status != types.ProposalStatusDeposit {
		return fmt.Errorf("proposal %d is not in deposit period (status: %s)", proposalId, proposal.StatusString())
	}

	// ── Validate depositor address ────────────────────────────────────────────
	depositorAddr, err := sdk.AccAddressFromBech32(depositor)
	if err != nil {
		return fmt.Errorf("invalid depositor address: %w", err)
	}

	// ── Lock deposit coins ────────────────────────────────────────────────────
	depositCoins := sdk.NewCoins(sdk.NewCoin(types.BondDenom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		depositorAddr,
		types.TreasuryModuleName,
		depositCoins,
	); err != nil {
		return fmt.Errorf("failed to lock deposit for proposal %d: %w", proposalId, err)
	}

	// ── Update TotalDeposit ───────────────────────────────────────────────────
	proposal.TotalDeposit = proposal.TotalDeposit.Add(amount)

	govParams := k.GetGovernanceParams(ctx)

	// ── Auto-advance to voting if threshold is now met ────────────────────────
	if proposal.TotalDeposit.GTE(govParams.MinDeposit) {
		proposal.Status = types.ProposalStatusVoting
		proposal.VotingEndTime = ctx.BlockHeight() + govParams.VotingPeriod

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalActivated,
				sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
				sdk.NewAttribute(types.AttributeKeyProposalStatus, proposal.StatusString()),
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			),
		)

		k.logger.Info("Proposal advanced to voting",
			"proposal_id", proposalId,
			"total_deposit", proposal.TotalDeposit.String(),
			"voting_end", proposal.VotingEndTime,
		)
	}

	if err := k.SetProposal(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal %d: %w", proposalId, err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDepositAdded,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor),
			sdk.NewAttribute(types.AttributeKeyDepositAmount, amount.String()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return nil
}

// ───────────────────────────────────────────────────────────────────────────────
// CastVote
// ───────────────────────────────────────────────────────────────────────────────

// CastVote records (or updates) a voter's choice on a proposal that is in VOTING
// status and within its voting period.
//
// Voting weight = sum of all active delegation amounts for the voter.
func (k Keeper) CastVote(ctx context.Context, proposalId uint64, voter string, option int32) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ── Validate option ───────────────────────────────────────────────────────
	if _, ok := types.VoteOptionName[option]; !ok {
		return fmt.Errorf("unknown vote option: %d", option)
	}

	// ── Validate voter address ────────────────────────────────────────────────
	if _, err := sdk.AccAddressFromBech32(voter); err != nil {
		return fmt.Errorf("invalid voter address: %w", err)
	}

	// ── Fetch proposal ────────────────────────────────────────────────────────
	proposal, found, err := k.GetProposal(ctx, proposalId)
	if err != nil {
		return fmt.Errorf("failed to retrieve proposal %d: %w", proposalId, err)
	}
	if !found {
		return fmt.Errorf("proposal %d not found", proposalId)
	}
	if proposal.Status != types.ProposalStatusVoting {
		return fmt.Errorf("proposal %d is not in voting period (status: %s)", proposalId, proposal.StatusString())
	}
	if sdkCtx.BlockHeight() > proposal.VotingEndTime {
		return fmt.Errorf("voting period for proposal %d has ended", proposalId)
	}

	// ── Compute voting weight from active delegations ─────────────────────────
	weight, err := k.getVoterStake(ctx, voter)
	if err != nil {
		return fmt.Errorf("failed to compute voting weight for %s: %w", voter, err)
	}
	if weight.IsZero() {
		return fmt.Errorf("voter %s has no staked VITA and cannot vote", voter)
	}

	// ── Upsert vote ───────────────────────────────────────────────────────────
	vote := types.Vote{
		ProposalId: proposalId,
		Voter:      voter,
		Option:     option,
		Weight:     weight,
		Timestamp:  sdkCtx.BlockTime().Unix(),
	}
	if err := k.SetVote(ctx, vote); err != nil {
		return fmt.Errorf("failed to store vote: %w", err)
	}

	// ── Emit event ────────────────────────────────────────────────────────────
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVoteCast,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter),
			sdk.NewAttribute(types.AttributeKeyVoteOption, types.VoteOptionName[option]),
			sdk.NewAttribute(types.AttributeKeyVoteWeight, weight.String()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	k.logger.Info("Vote cast",
		"proposal_id", proposalId,
		"voter", voter,
		"option", types.VoteOptionName[option],
		"weight", weight.String(),
	)
	return nil
}

// getVoterStake sums all active delegation amounts for the given voter address.
func (k Keeper) getVoterStake(ctx context.Context, voter string) (math.LegacyDec, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Build a prefix that covers all delegations by this delegator:
	// DelegationKeyPrefix | voterBytes
	delegatorPrefix := append(append([]byte{}, types.DelegationKeyPrefix...), []byte(voter)...)

	iter, err := store.Iterator(delegatorPrefix, storetypes.PrefixEndBytes(delegatorPrefix))
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("failed to open delegation iterator: %w", err)
	}
	defer iter.Close()

	total := math.LegacyZeroDec()
	for ; iter.Valid(); iter.Next() {
		// delegationRecord is unexported; decode just the Amount field via a local struct
		var rec struct {
			Amount string `json:"amount"`
		}
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			return math.LegacyZeroDec(), fmt.Errorf("failed to unmarshal delegation: %w", err)
		}
		amt, ok := math.NewIntFromString(rec.Amount)
		if !ok {
			return math.LegacyZeroDec(), fmt.Errorf("invalid delegation amount: %s", rec.Amount)
		}
		total = total.Add(math.LegacyNewDecFromInt(amt))
	}
	return total, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// TallyProposal
// ───────────────────────────────────────────────────────────────────────────────

// TallyProposal tallies all votes on a proposal, updates its status, and
// returns/burns its deposit accordingly.
//
// Tally rules:
//  1. participationRatio = totalVotingWeight / totalStaked < Quorum → REJECTED (low turnout)
//  2. vetoRatio = vetoVotes / totalVotingWeight >= VetoThreshold    → REJECTED + burn deposit
//  3. yesRatio = yesVotes / (yes+no+veto) >= Threshold             → PASSED
//  4. Otherwise                                                      → REJECTED
func (k Keeper) TallyProposal(ctx context.Context, proposalId uint64) (bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	proposal, found, err := k.GetProposal(ctx, proposalId)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve proposal %d: %w", proposalId, err)
	}
	if !found {
		return false, fmt.Errorf("proposal %d not found", proposalId)
	}
	if proposal.Status != types.ProposalStatusVoting {
		return false, fmt.Errorf("proposal %d is not in voting status", proposalId)
	}

	govParams := k.GetGovernanceParams(ctx)

	// ── Aggregate votes ───────────────────────────────────────────────────────
	votes, err := k.GetVotesByProposal(ctx, proposalId)
	if err != nil {
		return false, fmt.Errorf("failed to get votes for proposal %d: %w", proposalId, err)
	}

	yesVotes := math.LegacyZeroDec()
	noVotes := math.LegacyZeroDec()
	abstainVotes := math.LegacyZeroDec()
	vetoVotes := math.LegacyZeroDec()

	for _, v := range votes {
		switch v.Option {
		case types.VoteOptionYes:
			yesVotes = yesVotes.Add(v.Weight)
		case types.VoteOptionNo:
			noVotes = noVotes.Add(v.Weight)
		case types.VoteOptionAbstain:
			abstainVotes = abstainVotes.Add(v.Weight)
		case types.VoteOptionNoWithVeto:
			vetoVotes = vetoVotes.Add(v.Weight)
		}
	}

	totalVotingWeight := yesVotes.Add(noVotes).Add(abstainVotes).Add(vetoVotes)

	// ── Total staked VITA across all validators ───────────────────────────────
	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get validators: %w", err)
	}
	totalStaked := math.LegacyZeroDec()
	for _, val := range validators {
		totalStaked = totalStaked.Add(math.LegacyNewDecFromInt(val.TotalDelegated))
	}

	passed := false
	tallyReason := "rejected"
	burnDeposit := false

	if totalStaked.IsZero() || totalVotingWeight.IsZero() {
		tallyReason = "no_participation"
	} else {
		participationRatio := totalVotingWeight.Quo(totalStaked)

		if participationRatio.LT(govParams.Quorum) {
			tallyReason = "quorum_not_met"
		} else {
			vetoRatio := vetoVotes.Quo(totalVotingWeight)
			if vetoRatio.GTE(govParams.VetoThreshold) {
				tallyReason = "vetoed"
				burnDeposit = govParams.BurnDepositOnVeto
			} else {
				nonAbstain := yesVotes.Add(noVotes).Add(vetoVotes)
				if nonAbstain.IsPositive() {
					yesRatio := yesVotes.Quo(nonAbstain)
					if yesRatio.GTE(govParams.Threshold) {
						passed = true
						tallyReason = "passed"
					} else {
						tallyReason = "threshold_not_met"
					}
				} else {
					tallyReason = "no_non_abstain_votes"
				}
			}
		}
	}

	// ── Update proposal tally fields ──────────────────────────────────────────
	proposal.YesVotes = yesVotes
	proposal.NoVotes = noVotes
	proposal.AbstainVotes = abstainVotes
	proposal.VetoVotes = vetoVotes

	if passed {
		proposal.Status = types.ProposalStatusPassed
	} else {
		proposal.Status = types.ProposalStatusRejected
	}

	// ── Handle deposit ────────────────────────────────────────────────────────
	if proposal.TotalDeposit.IsPositive() {
		depositCoins := sdk.NewCoins(sdk.NewCoin(types.BondDenom, proposal.TotalDeposit))
		if burnDeposit {
			if err := k.bankKeeper.BurnCoins(ctx, types.TreasuryModuleName, depositCoins); err != nil {
				k.logger.Error("failed to burn vetoed deposit", "proposal_id", proposalId, "err", err)
			}
		} else {
			// Return deposit to proposer
			proposerAddr, err := sdk.AccAddressFromBech32(proposal.Proposer)
			if err == nil {
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryModuleName, proposerAddr, depositCoins); err != nil {
					k.logger.Error("failed to return deposit", "proposal_id", proposalId, "err", err)
				}
			}
		}
	}

	if err := k.SetProposal(ctx, proposal); err != nil {
		return false, fmt.Errorf("failed to update proposal %d after tally: %w", proposalId, err)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalTallied,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
			sdk.NewAttribute(types.AttributeKeyTallyPassed, fmt.Sprintf("%t", passed)),
			sdk.NewAttribute(types.AttributeKeyTallyReason, tallyReason),
			sdk.NewAttribute(types.AttributeKeyProposalStatus, proposal.StatusString()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	k.logger.Info("Proposal tallied",
		"proposal_id", proposalId,
		"passed", passed,
		"reason", tallyReason,
		"yes", yesVotes.String(),
		"no", noVotes.String(),
		"abstain", abstainVotes.String(),
		"veto", vetoVotes.String(),
	)

	return passed, nil
}

// ───────────────────────────────────────────────────────────────────────────────
// ExecuteProposal
// ───────────────────────────────────────────────────────────────────────────────

// ExecuteProposal runs the on-chain action encoded in a PASSED proposal.
//
// Supported types:
//   - "text"          : no-op
//   - "param_change"  : Content JSON {"key":"...","value":"..."} → UpdateParams
//   - "treasury_spend": Content JSON {"recipient":"...","amount":"...","denom":"..."}
//     → bankKeeper.SendCoinsFromModuleToAccount
func (k Keeper) ExecuteProposal(ctx context.Context, proposalId uint64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	proposal, found, err := k.GetProposal(ctx, proposalId)
	if err != nil {
		return fmt.Errorf("failed to retrieve proposal %d: %w", proposalId, err)
	}
	if !found {
		return fmt.Errorf("proposal %d not found", proposalId)
	}
	if proposal.Status != types.ProposalStatusPassed {
		return fmt.Errorf("proposal %d is not in PASSED status (status: %s)", proposalId, proposal.StatusString())
	}

	var execErr error

	switch proposal.ProposalType {
	case types.ProposalTypeText:
		// No-op execution for text proposals.

	case types.ProposalTypeParamChange:
		var payload struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(proposal.Content), &payload); err != nil {
			execErr = fmt.Errorf("invalid param_change content: %w", err)
		} else if payload.Key == "" {
			execErr = fmt.Errorf("param_change: key is empty")
		} else {
			if err := k.applyParamChange(ctx, payload.Key, payload.Value); err != nil {
				execErr = fmt.Errorf("param_change execution failed: %w", err)
			}
		}

	case types.ProposalTypeTreasurySpend:
		var payload struct {
			Recipient string `json:"recipient"`
			Amount    string `json:"amount"`
			Denom     string `json:"denom"`
		}
		if err := json.Unmarshal([]byte(proposal.Content), &payload); err != nil {
			execErr = fmt.Errorf("invalid treasury_spend content: %w", err)
		} else {
			recipientAddr, err := sdk.AccAddressFromBech32(payload.Recipient)
			if err != nil {
				execErr = fmt.Errorf("treasury_spend: invalid recipient: %w", err)
			} else {
				spendAmt, ok := math.NewIntFromString(payload.Amount)
				if !ok {
					execErr = fmt.Errorf("treasury_spend: invalid amount: %s", payload.Amount)
				} else {
					denom := payload.Denom
					if denom == "" {
						denom = types.BondDenom
					}
					coins := sdk.NewCoins(sdk.NewCoin(denom, spendAmt))
					if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryModuleName, recipientAddr, coins); err != nil {
						execErr = fmt.Errorf("treasury_spend transfer failed: %w", err)
					}
				}
			}
		}

	default:
		execErr = fmt.Errorf("unknown proposal type for execution: %s", proposal.ProposalType)
	}

	// ── Update status on failure ──────────────────────────────────────────────
	errStr := ""
	if execErr != nil {
		proposal.Status = types.ProposalStatusFailed
		if err := k.SetProposal(ctx, proposal); err != nil {
			k.logger.Error("failed to mark proposal as failed", "proposal_id", proposalId, "err", err)
		}
		errStr = execErr.Error()
		k.logger.Error("Proposal execution failed", "proposal_id", proposalId, "err", execErr)
	} else {
		k.logger.Info("Proposal executed successfully", "proposal_id", proposalId, "type", proposal.ProposalType)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalExecuted,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalId)),
			sdk.NewAttribute(types.AttributeKeyProposalType, proposal.ProposalType),
			sdk.NewAttribute(types.AttributeKeyProposalStatus, proposal.StatusString()),
			sdk.NewAttribute(types.AttributeKeyExecutionError, errStr),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return execErr
}

// applyParamChange dispatches a governance parameter change to the appropriate
// keeper method.  Only governance-specific params are handled here; fee / staking
// params remain governed by their own keepers.
func (k Keeper) applyParamChange(ctx context.Context, key, value string) error {
	switch key {
	case "min_deposit", "max_deposit_period", "voting_period", "quorum", "threshold", "veto_threshold":
		// Governance params are currently stored as defaults.
		// A future job will persist them; for now we log and succeed.
		k.logger.Info("param_change applied (gov params stored as defaults until Phase 5 Job 4)", "key", key, "value", value)
		return nil
	default:
		return fmt.Errorf("unknown governance param key: %s", key)
	}
}

// ───────────────────────────────────────────────────────────────────────────────
// EndBlockerGovernance
// ───────────────────────────────────────────────────────────────────────────────

// EndBlockerGovernance is called from the module EndBlocker on every block.
//
// It performs two sweeps:
//  1. Deposit-period proposals whose DepositEndTime has passed → FAILED (deposit refunded).
//  2. Voting-period proposals whose VotingEndTime has passed   → Tallied, then executed if passed.
func (k Keeper) EndBlockerGovernance(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentBlock := sdkCtx.BlockHeight()

	proposals, err := k.GetAllProposals(ctx)
	if err != nil {
		return fmt.Errorf("EndBlockerGovernance: failed to load proposals: %w", err)
	}

	for _, proposal := range proposals {
		switch proposal.Status {

		case types.ProposalStatusDeposit:
			if currentBlock <= proposal.DepositEndTime {
				continue
			}
			// Deposit period expired without reaching MinDeposit → FAIL + refund
			proposal.Status = types.ProposalStatusFailed
			if proposal.TotalDeposit.IsPositive() {
				depositCoins := sdk.NewCoins(sdk.NewCoin(types.BondDenom, proposal.TotalDeposit))
				proposerAddr, err := sdk.AccAddressFromBech32(proposal.Proposer)
				if err == nil {
					if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryModuleName, proposerAddr, depositCoins); err != nil {
						k.logger.Error("failed to refund deposit on expired proposal",
							"proposal_id", proposal.ProposalId, "err", err)
					}
				}
			}
			if err := k.SetProposal(ctx, proposal); err != nil {
				k.logger.Error("failed to mark deposit-expired proposal as failed",
					"proposal_id", proposal.ProposalId, "err", err)
				continue
			}
			k.logger.Info("Proposal deposit period expired — failed",
				"proposal_id", proposal.ProposalId)

		case types.ProposalStatusVoting:
			if currentBlock <= proposal.VotingEndTime {
				continue
			}
			// Voting period ended → tally
			passed, err := k.TallyProposal(ctx, proposal.ProposalId)
			if err != nil {
				k.logger.Error("failed to tally proposal",
					"proposal_id", proposal.ProposalId, "err", err)
				continue
			}
			// Execute if passed
			if passed {
				if err := k.ExecuteProposal(ctx, proposal.ProposalId); err != nil {
					k.logger.Error("failed to execute proposal",
						"proposal_id", proposal.ProposalId, "err", err)
				}
			}
		}
	}

	return nil
}
