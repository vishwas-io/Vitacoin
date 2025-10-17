package types

import (
	"fmt"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ===========================================================================================
// PHASE 3 - TASK 3.4: TREASURY TYPES
// ===========================================================================================
//
// Treasury-related type definitions for the VITACOIN module
//
// ===========================================================================================

// TreasurySpending represents a record of treasury expenditure
// All treasury spending is tracked for transparency and auditability
type TreasurySpending struct {
	// Unique identifier for this spending record
	Id string `json:"id"`
	
	// Governance proposal ID that authorized this spending
	ProposalId uint64 `json:"proposal_id"`
	
	// Recipient address that received the funds
	Recipient string `json:"recipient"`
	
	// Amount sent to recipient
	Amount sdk.Coins `json:"amount"`
	
	// Purpose of the spending (description from proposal)
	Purpose string `json:"purpose"`
	
	// Block height when spending occurred
	SpentHeight int64 `json:"spent_height"`
	
	// Unix timestamp when spending occurred
	SpentTime int64 `json:"spent_time"`
}

// Validate performs validation on TreasurySpending
func (ts TreasurySpending) Validate() error {
	if ts.Id == "" {
		return fmt.Errorf("treasury spending ID cannot be empty")
	}
	
	if _, err := sdk.AccAddressFromBech32(ts.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	
	if !ts.Amount.IsValid() || ts.Amount.IsZero() {
		return fmt.Errorf("invalid amount: %s", ts.Amount)
	}
	
	if ts.Purpose == "" {
		return fmt.Errorf("purpose cannot be empty")
	}
	
	if ts.SpentHeight < 0 {
		return fmt.Errorf("spent height cannot be negative")
	}
	
	if ts.SpentTime < 0 {
		return fmt.Errorf("spent time cannot be negative")
	}
	
	return nil
}

// String returns a human-readable string representation
func (ts TreasurySpending) String() string {
	return fmt.Sprintf(`TreasurySpending:
  ID:           %s
  Proposal ID:  %d
  Recipient:    %s
  Amount:       %s
  Purpose:      %s
  Spent Height: %d
  Spent Time:   %d`,
		ts.Id, ts.ProposalId, ts.Recipient, ts.Amount, 
		ts.Purpose, ts.SpentHeight, ts.SpentTime)
}

// Implement proto.Message interface for compatibility
func (ts *TreasurySpending) Reset()         {}
func (ts *TreasurySpending) ProtoMessage()  {}

// TreasuryStatistics represents comprehensive treasury statistics
type TreasuryStatistics struct {
	// Current balance in treasury module account
	CurrentBalance sdk.Coins `json:"current_balance"`
	
	// Total amount deposited to treasury (cumulative)
	TotalDeposited sdk.Coins `json:"total_deposited"`
	
	// Total amount spent from treasury (cumulative)
	TotalSpent sdk.Coins `json:"total_spent"`
	
	// Number of spending operations
	SpendingCount uint64 `json:"spending_count"`
	
	// Last update block height
	LastUpdateHeight int64 `json:"last_update_height"`
	
	// Last update unix timestamp
	LastUpdateTime int64 `json:"last_update_time"`
}

// Validate performs validation on TreasuryStatistics
func (ts TreasuryStatistics) Validate() error {
	if !ts.CurrentBalance.IsValid() {
		return fmt.Errorf("invalid current balance: %s", ts.CurrentBalance)
	}
	
	if !ts.TotalDeposited.IsValid() {
		return fmt.Errorf("invalid total deposited: %s", ts.TotalDeposited)
	}
	
	if !ts.TotalSpent.IsValid() {
		return fmt.Errorf("invalid total spent: %s", ts.TotalSpent)
	}
	
	if ts.LastUpdateHeight < 0 {
		return fmt.Errorf("last update height cannot be negative")
	}
	
	return nil
}

// String returns a human-readable string representation
func (ts TreasuryStatistics) String() string {
	return fmt.Sprintf(`TreasuryStatistics:
  Current Balance:   %s
  Total Deposited:   %s
  Total Spent:       %s
  Spending Count:    %d
  Last Update:       height=%d, time=%d`,
		ts.CurrentBalance, ts.TotalDeposited, ts.TotalSpent,
		ts.SpendingCount, ts.LastUpdateHeight, ts.LastUpdateTime)
}

// TreasuryGenesisState represents treasury state for genesis import/export
type TreasuryGenesisState struct {
	// Current treasury balance
	Balance sdk.Coins `json:"balance"`
	
	// Historical spending records
	SpendingList []TreasurySpending `json:"spending_list"`
}

// Validate performs validation on TreasuryGenesisState
func (tgs TreasuryGenesisState) Validate() error {
	if !tgs.Balance.IsValid() {
		return fmt.Errorf("invalid treasury balance: %s", tgs.Balance)
	}
	
	for i, spending := range tgs.SpendingList {
		if err := spending.Validate(); err != nil {
			return fmt.Errorf("invalid spending record at index %d: %w", i, err)
		}
	}
	
	return nil
}

// DefaultTreasuryGenesisState returns default treasury genesis state
func DefaultTreasuryGenesisState() *TreasuryGenesisState {
	return &TreasuryGenesisState{
		Balance:      sdk.NewCoins(),
		SpendingList: []TreasurySpending{},
	}
}

// TreasurySpendProposal represents a governance proposal to spend from treasury
// This will be used for governance integration
type TreasurySpendProposal struct {
	// Title of the proposal
	Title string `json:"title"`
	
	// Description of what the funds will be used for
	Description string `json:"description"`
	
	// Recipient address
	Recipient string `json:"recipient"`
	
	// Amount to spend
	Amount sdk.Coins `json:"amount"`
}

// ValidateBasic performs basic validation
func (tsp TreasurySpendProposal) ValidateBasic() error {
	if tsp.Title == "" {
		return fmt.Errorf("proposal title cannot be empty")
	}
	
	if tsp.Description == "" {
		return fmt.Errorf("proposal description cannot be empty")
	}
	
	if _, err := sdk.AccAddressFromBech32(tsp.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	
	if !tsp.Amount.IsValid() || tsp.Amount.IsZero() {
		return fmt.Errorf("invalid amount: %s", tsp.Amount)
	}
	
	return nil
}

// String returns a human-readable string representation
func (tsp TreasurySpendProposal) String() string {
	return fmt.Sprintf(`TreasurySpendProposal:
  Title:       %s
  Description: %s
  Recipient:   %s
  Amount:      %s`,
		tsp.Title, tsp.Description, tsp.Recipient, tsp.Amount)
}

// ProposalRoute returns the routing key for the proposal
func (tsp TreasurySpendProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns the type of the proposal
func (tsp TreasurySpendProposal) ProposalType() string {
	return "TreasurySpend"
}

// TreasuryImpactEstimate represents the estimated impact of a treasury spending
type TreasuryImpactEstimate struct {
	CurrentBalance sdk.Coins `json:"current_balance"`
	SpendAmount    sdk.Coins `json:"spend_amount"`
	NewBalance     sdk.Coins `json:"new_balance"`
	CurrentRunway  int64     `json:"current_runway"`
	NewRunway      int64     `json:"new_runway"`
	CurrentHealth  uint32    `json:"current_health"`
	NewHealth      uint32    `json:"new_health"`
	Recommended    bool      `json:"recommended"`
}

// TreasurySpendingReport represents a comprehensive treasury activity report
type TreasurySpendingReport struct {
	FromHeight      int64                  `json:"from_height"`
	ToHeight        int64                  `json:"to_height"`
	TotalSpent      sdk.Coins              `json:"total_spent"`
	SpendingCount   uint64                 `json:"spending_count"`
	SpendingRecords []TreasurySpending     `json:"spending_records"`
	ByRecipient     map[string]sdk.Coins   `json:"by_recipient"`
	ByPurpose       map[string]sdk.Coins   `json:"by_purpose"`
	CurrentStats    *TreasuryStatistics    `json:"current_stats"`
}
