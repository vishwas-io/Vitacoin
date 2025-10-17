package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ===========================================================================================
// PHASE 3 - TASK 3.4: TREASURY QUERY TYPES
// ===========================================================================================
//
// Query request and response types for treasury gRPC endpoints
// These are temporary Go implementations pending proto file generation
//
// ===========================================================================================

// QueryTreasuryBalanceRequest is the request type for the Query/TreasuryBalance RPC method
type QueryTreasuryBalanceRequest struct{}

// QueryTreasuryBalanceResponse is the response type for the Query/TreasuryBalance RPC method
type QueryTreasuryBalanceResponse struct {
	Balance sdk.Coins `json:"balance"`
}

// QueryTreasuryStatisticsRequest is the request type for the Query/TreasuryStatistics RPC method
type QueryTreasuryStatisticsRequest struct{}

// QueryTreasuryStatisticsResponse is the response type for the Query/TreasuryStatistics RPC method
type QueryTreasuryStatisticsResponse struct {
	Statistics *TreasuryStatistics `json:"statistics"`
}

// QueryTreasurySpendingRequest is the request type for the Query/TreasurySpending RPC method
type QueryTreasurySpendingRequest struct {
	Id string `json:"id"`
}

// QueryTreasurySpendingResponse is the response type for the Query/TreasurySpending RPC method
type QueryTreasurySpendingResponse struct {
	Spending TreasurySpending `json:"spending"`
}

// QueryTreasurySpendingAllRequest is the request type for the Query/TreasurySpendingAll RPC method
type QueryTreasurySpendingAllRequest struct{}

// QueryTreasurySpendingAllResponse is the response type for the Query/TreasurySpendingAll RPC method
type QueryTreasurySpendingAllResponse struct {
	Spending []TreasurySpending `json:"spending"`
}

// QueryTreasurySpendingByProposalRequest is the request type for the Query/TreasurySpendingByProposal RPC method
type QueryTreasurySpendingByProposalRequest struct {
	ProposalId uint64 `json:"proposal_id"`
}

// QueryTreasurySpendingByProposalResponse is the response type for the Query/TreasurySpendingByProposal RPC method
type QueryTreasurySpendingByProposalResponse struct {
	Spending []TreasurySpending `json:"spending"`
}

// QueryTreasurySpendingByRecipientRequest is the request type for the Query/TreasurySpendingByRecipient RPC method
type QueryTreasurySpendingByRecipientRequest struct {
	Recipient string `json:"recipient"`
}

// QueryTreasurySpendingByRecipientResponse is the response type for the Query/TreasurySpendingByRecipient RPC method
type QueryTreasurySpendingByRecipientResponse struct {
	Spending []TreasurySpending `json:"spending"`
}

// QueryTreasurySpendingReportRequest is the request type for the Query/TreasurySpendingReport RPC method
type QueryTreasurySpendingReportRequest struct {
	FromHeight int64 `json:"from_height"`
	ToHeight   int64 `json:"to_height"`
}

// QueryTreasurySpendingReportResponse is the response type for the Query/TreasurySpendingReport RPC method
type QueryTreasurySpendingReportResponse struct {
	Report *TreasurySpendingReport `json:"report"`
}

// QueryTreasuryHealthRequest is the request type for the Query/TreasuryHealth RPC method
type QueryTreasuryHealthRequest struct{}

// QueryTreasuryHealthResponse is the response type for the Query/TreasuryHealth RPC method
type QueryTreasuryHealthResponse struct {
	Balance     math.Int `json:"balance"`
	Runway      int64    `json:"runway"`
	HealthScore uint32   `json:"health_score"`
}

// QueryTreasuryImpactEstimateRequest is the request type for the Query/TreasuryImpactEstimate RPC method
type QueryTreasuryImpactEstimateRequest struct {
	Amount sdk.Coins `json:"amount"`
}

// QueryTreasuryImpactEstimateResponse is the response type for the Query/TreasuryImpactEstimate RPC method
type QueryTreasuryImpactEstimateResponse struct {
	Estimate *TreasuryImpactEstimate `json:"estimate"`
}
