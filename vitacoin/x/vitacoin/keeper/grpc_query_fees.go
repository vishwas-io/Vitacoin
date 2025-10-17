package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// Phase 3: Fee & Economics Query Implementations

// FeeStatistics implements the Query/FeeStatistics gRPC method
// Returns cumulative fee statistics since genesis
func (q queryServer) FeeStatistics(ctx context.Context, req *types.QueryFeeStatisticsRequest) (*types.QueryFeeStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Get fee statistics from keeper
	stats, err := q.Keeper.GetFeeStatistics(ctx)
	if err != nil {
		// If no statistics found (e.g., genesis state), return zeros
		q.Keeper.Logger().Debug("fee statistics not found, returning defaults", "error", err)
		
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		return &types.QueryFeeStatisticsResponse{
			TotalCollectedAllTime:    math.ZeroInt(),
			TotalBurnedAllTime:       math.ZeroInt(),
			TotalToValidatorsAllTime: math.ZeroInt(),
			TotalToTreasuryAllTime:   math.ZeroInt(),
			TotalTransactionsAllTime: 0,
			LastUpdateHeight:         sdkCtx.BlockHeight(),
			CurrentEpoch:             q.Keeper.CalculateEpoch(ctx),
		}, nil
	}

	// Log query for monitoring
	q.Keeper.Logger().Debug("fee statistics query", 
		"total_collected", stats.TotalCollectedAllTime.String(),
		"total_burned", stats.TotalBurnedAllTime.String(),
		"total_transactions", stats.TotalTransactionsAllTime,
	)

	return &types.QueryFeeStatisticsResponse{
		TotalCollectedAllTime:    stats.TotalCollectedAllTime,
		TotalBurnedAllTime:       stats.TotalBurnedAllTime,
		TotalToValidatorsAllTime: stats.TotalToValidatorsAllTime,
		TotalToTreasuryAllTime:   stats.TotalToTreasuryAllTime,
		TotalTransactionsAllTime: stats.TotalTransactionsAllTime,
		LastUpdateHeight:         stats.LastUpdateHeight,
		CurrentEpoch:             stats.CurrentEpoch,
	}, nil
}

// BurnStatistics implements the Query/BurnStatistics gRPC method
// Returns burn mechanism statistics and supply tracking
func (q queryServer) BurnStatistics(ctx context.Context, req *types.QueryBurnStatisticsRequest) (*types.QueryBurnStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Get burn statistics from keeper
	stats, err := q.Keeper.GetBurnStatistics(ctx)
	if err != nil {
		// If no burn statistics found (e.g., genesis state), return defaults
		q.Keeper.Logger().Debug("burn statistics not found, returning defaults", "error", err)
		
		// Get params for burn cap
		params, paramErr := q.Keeper.GetParams(ctx)
		if paramErr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get params: %s", paramErr.Error()))
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		initialSupply := math.NewInt(1000000000).Mul(math.NewInt(1000000000000000000)) // 1B VITA
		
		return &types.QueryBurnStatisticsResponse{
			TotalBurned:     math.ZeroInt(),
			BurnRatePerDay:  math.ZeroInt(),
			CurrentSupply:   initialSupply,
			BurnCapSupply:   params.BurnCapSupply,
			RemainingToCap:  initialSupply.Sub(params.BurnCapSupply),
			BurnCapReached:  false,
			LastBurnHeight:  sdkCtx.BlockHeight(),
		}, nil
	}

	// Log query for monitoring
	q.Keeper.Logger().Debug("burn statistics query",
		"total_burned", stats.TotalBurned.String(),
		"current_supply", stats.CurrentSupply.String(),
		"burn_cap_reached", stats.BurnCapReached,
	)

	return &types.QueryBurnStatisticsResponse{
		TotalBurned:     stats.TotalBurned,
		BurnRatePerDay:  stats.BurnRatePerDay,
		CurrentSupply:   stats.CurrentSupply,
		BurnCapSupply:   stats.BurnCapSupply,
		RemainingToCap:  stats.RemainingToCap,
		BurnCapReached:  stats.BurnCapReached,
		LastBurnHeight:  stats.LastBurnHeight,
	}, nil
}

// SupplySnapshot implements the Query/SupplySnapshot gRPC method
// Returns supply snapshot for a specific block height
func (q queryServer) SupplySnapshot(ctx context.Context, req *types.QuerySupplySnapshotRequest) (*types.QuerySupplySnapshotResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Validate height
	if req.Height <= 0 {
		return nil, status.Error(codes.InvalidArgument, "height must be positive")
	}

	// Get current height
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentHeight := sdkCtx.BlockHeight()
	
	// Validate height is not in the future
	if req.Height > currentHeight {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("height %d is in the future (current: %d)", req.Height, currentHeight))
	}

	// Get supply snapshot from keeper
	snapshot, err := q.Keeper.GetSupplySnapshot(ctx, req.Height)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("supply snapshot not found for height %d: %s", req.Height, err.Error()))
	}

	// Log query for monitoring
	q.Keeper.Logger().Debug("supply snapshot query",
		"height", snapshot.Height,
		"total_supply", snapshot.TotalSupply.String(),
		"circulating_supply", snapshot.CirculatingSupply.String(),
	)

	return &types.QuerySupplySnapshotResponse{
		Height:            snapshot.Height,
		Timestamp:         snapshot.Timestamp.Unix(),
		TotalSupply:       snapshot.TotalSupply,
		CirculatingSupply: snapshot.CirculatingSupply,
		LiquidSupply:      snapshot.LiquidSupply,
		BondedSupply:      snapshot.BondedSupply,
		BurnedCumulative:  snapshot.BurnedCumulative,
	}, nil
}

// SupplySnapshotLatest implements the Query/SupplySnapshotLatest gRPC method
// Returns the most recent supply snapshot
func (q queryServer) SupplySnapshotLatest(ctx context.Context, req *types.QuerySupplySnapshotLatestRequest) (*types.QuerySupplySnapshotLatestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Get latest supply snapshot from keeper
	snapshot, err := q.Keeper.GetLatestSupplySnapshot(ctx)
	if err != nil {
		// If no snapshot found, create one on-the-fly
		q.Keeper.Logger().Debug("no supply snapshot found, creating current snapshot", "error", err)
		
		createErr := q.Keeper.CreateSupplySnapshot(ctx)
		if createErr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create supply snapshot: %s", createErr.Error()))
		}
		
		// Try to get it again
		snapshot, err = q.Keeper.GetLatestSupplySnapshot(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to retrieve created snapshot: %s", err.Error()))
		}
	}

	// Log query for monitoring
	q.Keeper.Logger().Debug("latest supply snapshot query",
		"height", snapshot.Height,
		"total_supply", snapshot.TotalSupply.String(),
		"circulating_supply", snapshot.CirculatingSupply.String(),
	)

	return &types.QuerySupplySnapshotLatestResponse{
		Height:            snapshot.Height,
		Timestamp:         snapshot.Timestamp.Unix(),
		TotalSupply:       snapshot.TotalSupply,
		CirculatingSupply: snapshot.CirculatingSupply,
		LiquidSupply:      snapshot.LiquidSupply,
		BondedSupply:      snapshot.BondedSupply,
		BurnedCumulative:  snapshot.BurnedCumulative,
	}, nil
}

// FeeAccumulator implements the Query/FeeAccumulator gRPC method
// Returns the current block's fee accumulator (in-progress fees)
func (q queryServer) FeeAccumulator(ctx context.Context, req *types.QueryFeeAccumulatorRequest) (*types.QueryFeeAccumulatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Get current block fee accumulator
	accumulator, err := q.Keeper.GetBlockFeeAccumulator(ctx)
	if err != nil {
		// If no accumulator found (e.g., beginning of block), return zeros
		q.Keeper.Logger().Debug("fee accumulator not found, returning defaults", "error", err)
		
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		return &types.QueryFeeAccumulatorResponse{
			Height:           sdkCtx.BlockHeight(),
			TotalCollected:   math.ZeroInt(),
			TransactionCount: 0,
		}, nil
	}

	// Log query for monitoring
	q.Keeper.Logger().Debug("fee accumulator query",
		"height", accumulator.Height,
		"total_collected", accumulator.TotalCollected.String(),
		"transaction_count", accumulator.TransactionCount,
	)

	return &types.QueryFeeAccumulatorResponse{
		Height:           accumulator.Height,
		TotalCollected:   accumulator.TotalCollected,
		TransactionCount: accumulator.TransactionCount,
	}, nil
}
