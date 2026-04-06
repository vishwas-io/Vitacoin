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
