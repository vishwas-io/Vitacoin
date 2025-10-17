package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Params implements the Query/Params gRPC method
// Returns the module parameters
func (q queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := q.Keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Merchant implements the Query/Merchant gRPC method
// Returns a specific merchant by address
func (q queryServer) Merchant(ctx context.Context, req *types.QueryMerchantRequest) (*types.QueryMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant address cannot be empty")
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid address: %s", err.Error()))
	}

	merchant, err := q.Keeper.GetMerchant(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("merchant not found: %s", err.Error()))
	}

	return &types.QueryMerchantResponse{Merchant: merchant}, nil
}

// MerchantAll implements the Query/MerchantAll gRPC method
// Returns all merchants with pagination support
func (q queryServer) MerchantAll(ctx context.Context, req *types.QueryAllMerchantRequest) (*types.QueryAllMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	merchants, err := q.Keeper.GetAllMerchants(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMerchantResponse{
		Merchants:  merchants,
		Pagination: nil, // TODO: Implement pagination in Phase 3
	}, nil
}

// Payment implements the Query/Payment gRPC method
// Returns a specific payment by ID
func (q queryServer) Payment(ctx context.Context, req *types.QueryPaymentRequest) (*types.QueryPaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "payment ID cannot be empty")
	}

	payment, err := q.Keeper.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("payment not found: %s", err.Error()))
	}

	return &types.QueryPaymentResponse{Payment: payment}, nil
}

// PaymentAll implements the Query/PaymentAll gRPC method
// Returns all payments with pagination support
func (q queryServer) PaymentAll(ctx context.Context, req *types.QueryAllPaymentRequest) (*types.QueryAllPaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	payments, err := q.Keeper.GetAllPayments(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPaymentResponse{
		Payments:   payments,
		Pagination: nil, // TODO: Implement pagination in Phase 3
	}, nil
}

// Vault implements the Query/Vault gRPC method
// Returns a specific vault by ID
func (q queryServer) Vault(ctx context.Context, req *types.QueryVaultRequest) (*types.QueryVaultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "vault ID cannot be empty")
	}

	vault, err := q.Keeper.GetVault(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("vault not found: %s", err.Error()))
	}

	return &types.QueryVaultResponse{Vault: vault}, nil
}

// VaultAll implements the Query/VaultAll gRPC method
// Returns all vaults with pagination support
func (q queryServer) VaultAll(ctx context.Context, req *types.QueryAllVaultRequest) (*types.QueryAllVaultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	vaults, err := q.Keeper.GetAllVaults(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllVaultResponse{
		Vaults:     vaults,
		Pagination: nil, // TODO: Implement pagination in Phase 3
	}, nil
}

// RewardPool implements the Query/RewardPool gRPC method
// Returns a specific reward pool by ID
func (q queryServer) RewardPool(ctx context.Context, req *types.QueryRewardPoolRequest) (*types.QueryRewardPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "reward pool ID cannot be empty")
	}

	pool, err := q.Keeper.GetRewardPool(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("reward pool not found: %s", err.Error()))
	}

	return &types.QueryRewardPoolResponse{Pool: pool}, nil
}

// RewardPoolAll implements the Query/RewardPoolAll gRPC method
// Returns all reward pools with pagination support
func (q queryServer) RewardPoolAll(ctx context.Context, req *types.QueryAllRewardPoolRequest) (*types.QueryAllRewardPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	pools, err := q.Keeper.GetAllRewardPools(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRewardPoolResponse{
		Pools:      pools,
		Pagination: nil, // TODO: Implement pagination in Phase 3
	}, nil
}
