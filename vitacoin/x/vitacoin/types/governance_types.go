package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// ===========================================================================================
// PHASE 5 - GOVERNANCE TYPES
// ===========================================================================================

// ProposalStatus enum — matches Cosmos Gov v1 conventions
const (
	ProposalStatusNil      int32 = 0
	ProposalStatusDeposit  int32 = 1
	ProposalStatusVoting   int32 = 2
	ProposalStatusPassed   int32 = 3
	ProposalStatusRejected int32 = 4
	ProposalStatusFailed   int32 = 5
)

// ProposalStatusName maps status codes to human-readable strings.
var ProposalStatusName = map[int32]string{
	ProposalStatusNil:      "NIL",
	ProposalStatusDeposit:  "DEPOSIT",
	ProposalStatusVoting:   "VOTING",
	ProposalStatusPassed:   "PASSED",
	ProposalStatusRejected: "REJECTED",
	ProposalStatusFailed:   "FAILED",
}

// VoteOption enum
const (
	VoteOptionYes        int32 = 1
	VoteOptionAbstain    int32 = 2
	VoteOptionNo         int32 = 3
	VoteOptionNoWithVeto int32 = 4
)

// VoteOptionName maps vote options to human-readable strings.
var VoteOptionName = map[int32]string{
	VoteOptionYes:        "YES",
	VoteOptionAbstain:    "ABSTAIN",
	VoteOptionNo:         "NO",
	VoteOptionNoWithVeto: "NO_WITH_VETO",
}

// ProposalType constants
const (
	ProposalTypeText          = "text"
	ProposalTypeParamChange   = "param_change"
	ProposalTypeTreasurySpend = "treasury_spend"
)

// Proposal represents an on-chain governance proposal.
type Proposal struct {
	// Unique sequential proposal ID.
	ProposalId uint64 `json:"proposal_id"`

	// Short human-readable title.
	Title string `json:"title"`

	// Full description / motivation.
	Description string `json:"description"`

	// Type discriminates proposal content: "text", "param_change", "treasury_spend".
	ProposalType string `json:"proposal_type"`

	// Current lifecycle status.
	Status int32 `json:"status"`

	// Bech32 address of the proposer.
	Proposer string `json:"proposer"`

	// Unix timestamp (seconds) when proposal was submitted.
	SubmitTime int64 `json:"submit_time"`

	// Block height after which deposit period ends and proposal fails if underfunded.
	DepositEndTime int64 `json:"deposit_end_time"`

	// Block height after which voting is tallied.
	VotingEndTime int64 `json:"voting_end_time"`

	// Total VITA (in avita) currently deposited for this proposal.
	TotalDeposit math.Int `json:"total_deposit"`

	// Tallied vote weights (in staked VITA, LegacyDec for fractional precision).
	YesVotes     math.LegacyDec `json:"yes_votes"`
	NoVotes      math.LegacyDec `json:"no_votes"`
	AbstainVotes math.LegacyDec `json:"abstain_votes"`
	VetoVotes    math.LegacyDec `json:"veto_votes"`

	// JSON-encoded type-specific proposal payload.
	Content string `json:"content"`
}

// Validate performs stateless validation on a Proposal.
func (p Proposal) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("proposal title cannot be empty")
	}
	if p.Description == "" {
		return fmt.Errorf("proposal description cannot be empty")
	}
	if p.ProposalType != ProposalTypeText &&
		p.ProposalType != ProposalTypeParamChange &&
		p.ProposalType != ProposalTypeTreasurySpend {
		return fmt.Errorf("unknown proposal type: %s", p.ProposalType)
	}
	if p.Proposer == "" {
		return fmt.Errorf("proposer address cannot be empty")
	}
	if p.TotalDeposit.IsNil() {
		return fmt.Errorf("total deposit must be initialized")
	}
	if p.TotalDeposit.IsNegative() {
		return fmt.Errorf("total deposit cannot be negative")
	}
	return nil
}

// StatusString returns a human-readable status string.
func (p Proposal) StatusString() string {
	if name, ok := ProposalStatusName[p.Status]; ok {
		return name
	}
	return "UNKNOWN"
}

// Vote represents a single voter's choice on a proposal.
type Vote struct {
	// Proposal being voted on.
	ProposalId uint64 `json:"proposal_id"`

	// Bech32 address of the voter.
	Voter string `json:"voter"`

	// Chosen vote option.
	Option int32 `json:"option"`

	// Voting power at time of vote (staked VITA, LegacyDec for sub-unit precision).
	Weight math.LegacyDec `json:"weight"`

	// Unix timestamp (seconds) when the vote was cast.
	Timestamp int64 `json:"timestamp"`
}

// Validate performs stateless validation on a Vote.
func (v Vote) Validate() error {
	if v.Voter == "" {
		return fmt.Errorf("voter address cannot be empty")
	}
	if _, ok := VoteOptionName[v.Option]; !ok {
		return fmt.Errorf("unknown vote option: %d", v.Option)
	}
	if v.Weight.IsNil() {
		return fmt.Errorf("vote weight must be initialized")
	}
	if v.Weight.IsNegative() {
		return fmt.Errorf("vote weight cannot be negative")
	}
	return nil
}

// GovernanceParams holds the tunable governance parameters.
// These are read from the store; DefaultGovernanceParams is used when no record exists.
type GovernanceParams struct {
	// Minimum VITA deposit (in avita) required to move a proposal into voting.
	MinDeposit math.Int `json:"min_deposit"`

	// Number of blocks after submission during which deposits may be added.
	// Proposal is rejected if MinDeposit is not met before this deadline.
	MaxDepositPeriod int64 `json:"max_deposit_period"`

	// Number of blocks the proposal is open for voting once it enters the voting period.
	VotingPeriod int64 `json:"voting_period"`

	// Fraction of staked VITA that must participate for the tally to be valid.
	// e.g., 0.334 → 33.4 % quorum required.
	Quorum math.LegacyDec `json:"quorum"`

	// Fraction of non-abstain votes that must be YES for a proposal to pass.
	// e.g., 0.5 → simple majority.
	Threshold math.LegacyDec `json:"threshold"`

	// Fraction of total votes that must be NoWithVeto for the proposal to be vetoed.
	// e.g., 0.334 → 33.4 % veto threshold.
	VetoThreshold math.LegacyDec `json:"veto_threshold"`

	// If true, deposits are burned when a proposal is vetoed.
	BurnDepositOnVeto bool `json:"burn_deposit_on_veto"`
}

// Validate performs stateless validation on GovernanceParams.
func (gp GovernanceParams) Validate() error {
	if gp.MinDeposit.IsNil() || gp.MinDeposit.IsNegative() {
		return fmt.Errorf("min deposit must be a non-negative value")
	}
	if gp.MaxDepositPeriod <= 0 {
		return fmt.Errorf("max deposit period must be positive")
	}
	if gp.VotingPeriod <= 0 {
		return fmt.Errorf("voting period must be positive")
	}
	if gp.Quorum.IsNil() || gp.Quorum.IsNegative() || gp.Quorum.GT(math.LegacyOneDec()) {
		return fmt.Errorf("quorum must be between 0 and 1")
	}
	if gp.Threshold.IsNil() || gp.Threshold.IsNegative() || gp.Threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("threshold must be between 0 and 1")
	}
	if gp.VetoThreshold.IsNil() || gp.VetoThreshold.IsNegative() || gp.VetoThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("veto threshold must be between 0 and 1")
	}
	return nil
}

// DefaultGovernanceParams returns the default (hardcoded) governance parameters.
//
//	MinDeposit       = 10,000 VITA  (= 10_000 * 10^18 avita)
//	MaxDepositPeriod = 14,400 blocks (~1 day at 6 s/block)
//	VotingPeriod     = 100,800 blocks (~7 days at 6 s/block)
//	Quorum           = 33.4 %
//	Threshold        = 50 %
//	VetoThreshold    = 33.4 %
//	BurnDepositOnVeto = true
func DefaultGovernanceParams() GovernanceParams {
	// 10_000 VITA in avita (18 decimal places)
	minDeposit, _ := math.NewIntFromString("10000000000000000000000") // 10_000 * 10^18

	return GovernanceParams{
		MinDeposit:        minDeposit,
		MaxDepositPeriod:  14400,   // ~1 day
		VotingPeriod:      100800,  // ~7 days
		Quorum:            math.LegacyMustNewDecFromStr("0.334"),
		Threshold:         math.LegacyMustNewDecFromStr("0.500"),
		VetoThreshold:     math.LegacyMustNewDecFromStr("0.334"),
		BurnDepositOnVeto: true,
	}
}
