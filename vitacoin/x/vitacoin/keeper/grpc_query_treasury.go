package keeper

import (
	"context"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ===========================================================================================
// PHASE 3 - TASK 3.4: TREASURY QUERY ENDPOINTS
// ===========================================================================================
//
// gRPC query implementations for treasury operations
// Provides comprehensive treasury state and statistics queries
//
// ===========================================================================================

// TreasuryBalance queries the current treasury balance
func (k Keeper) TreasuryBalance(goCtx context.Context, req *types.QueryTreasuryBalanceRequest) (*types.QueryTreasuryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	balance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasuryBalanceResponse{
		Balance: balance,
	}, nil
}

// TreasuryStatistics queries comprehensive treasury statistics
func (k Keeper) TreasuryStatistics(goCtx context.Context, req *types.QueryTreasuryStatisticsRequest) (*types.QueryTreasuryStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	stats, err := k.GetTreasuryStatistics(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasuryStatisticsResponse{
		Statistics: stats,
	}, nil
}

// TreasurySpending queries a specific treasury spending record
func (k Keeper) TreasurySpending(goCtx context.Context, req *types.QueryTreasurySpendingRequest) (*types.QueryTreasurySpendingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "spending ID cannot be empty")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	spending, err := k.GetTreasurySpending(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	
	return &types.QueryTreasurySpendingResponse{
		Spending: spending,
	}, nil
}

// TreasurySpendingAll queries all treasury spending records
func (k Keeper) TreasurySpendingAll(goCtx context.Context, req *types.QueryTreasurySpendingAllRequest) (*types.QueryTreasurySpendingAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	spending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasurySpendingAllResponse{
		Spending: spending,
	}, nil
}

// TreasurySpendingByProposal queries treasury spending for a specific proposal
func (k Keeper) TreasurySpendingByProposal(goCtx context.Context, req *types.QueryTreasurySpendingByProposalRequest) (*types.QueryTreasurySpendingByProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	spending, err := k.GetTreasurySpendingByProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasurySpendingByProposalResponse{
		Spending: spending,
	}, nil
}

// TreasurySpendingByRecipient queries treasury spending for a specific recipient
func (k Keeper) TreasurySpendingByRecipient(goCtx context.Context, req *types.QueryTreasurySpendingByRecipientRequest) (*types.QueryTreasurySpendingByRecipientResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	if req.Recipient == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient address cannot be empty")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	spending, err := k.GetTreasurySpendingByRecipient(ctx, req.Recipient)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasurySpendingByRecipientResponse{
		Spending: spending,
	}, nil
}

// TreasurySpendingReport queries treasury spending report for a height range
func (k Keeper) TreasurySpendingReport(goCtx context.Context, req *types.QueryTreasurySpendingReportRequest) (*types.QueryTreasurySpendingReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	if req.FromHeight < 0 {
		return nil, status.Error(codes.InvalidArgument, "from_height cannot be negative")
	}
	
	if req.ToHeight < 0 {
		return nil, status.Error(codes.InvalidArgument, "to_height cannot be negative")
	}
	
	if req.ToHeight < req.FromHeight {
		return nil, status.Error(codes.InvalidArgument, "to_height must be greater than or equal to from_height")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	report, err := k.GetTreasurySpendingReport(ctx, req.FromHeight, req.ToHeight)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasurySpendingReportResponse{
		Report: report,
	}, nil
}

// TreasuryHealth queries the current treasury health score
func (k Keeper) TreasuryHealth(goCtx context.Context, req *types.QueryTreasuryHealthRequest) (*types.QueryTreasuryHealthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	balance, err := k.GetVitaTreasuryBalance(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	runway, err := k.EstimateTreasuryRunway(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	healthScore, err := k.GetTreasuryHealth(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasuryHealthResponse{
		Balance:     balance,
		Runway:      runway,
		HealthScore: healthScore,
	}, nil
}

// TreasuryImpactEstimate queries the estimated impact of a proposed spending
func (k Keeper) TreasuryImpactEstimate(goCtx context.Context, req *types.QueryTreasuryImpactEstimateRequest) (*types.QueryTreasuryImpactEstimateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	if !req.Amount.IsValid() || req.Amount.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	estimate, err := k.EstimateTreasurySpendImpact(ctx, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryTreasuryImpactEstimateResponse{
		Estimate: estimate,
	}, nil
}
