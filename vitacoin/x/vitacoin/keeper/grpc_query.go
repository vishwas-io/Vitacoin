package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/runtime"

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

	kvStore := q.Keeper.storeService.OpenKVStore(ctx)
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(kvStore), types.MerchantKeyPrefix)

	var merchants []types.Merchant
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var merchant types.Merchant
		if err := q.Keeper.cdc.Unmarshal(value, &merchant); err != nil {
			return err
		}
		merchants = append(merchants, merchant)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMerchantResponse{
		Merchants:  merchants,
		Pagination: pageRes,
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

	kvStore := q.Keeper.storeService.OpenKVStore(ctx)
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(kvStore), types.PaymentKeyPrefix)

	var payments []types.Payment
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var payment types.Payment
		if err := q.Keeper.cdc.Unmarshal(value, &payment); err != nil {
			return err
		}
		payments = append(payments, payment)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPaymentResponse{
		Payments:   payments,
		Pagination: pageRes,
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

	kvStore := q.Keeper.storeService.OpenKVStore(ctx)
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(kvStore), types.VaultKeyPrefix)

	var vaults []types.Vault
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var vault types.Vault
		if err := q.Keeper.cdc.Unmarshal(value, &vault); err != nil {
			return err
		}
		vaults = append(vaults, vault)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllVaultResponse{
		Vaults:     vaults,
		Pagination: pageRes,
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

	kvStore := q.Keeper.storeService.OpenKVStore(ctx)
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(kvStore), types.RewardPoolKeyPrefix)

	var pools []types.RewardPool
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		var pool types.RewardPool
		if err := q.Keeper.cdc.Unmarshal(value, &pool); err != nil {
			return err
		}
		pools = append(pools, pool)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRewardPoolResponse{
		Pools:      pools,
		Pagination: pageRes,
	}, nil
}
