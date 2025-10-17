package types

import (
	"cosmossdk.io/math"
)

// Phase 3: Fee Query Request/Response Types
// These are manually created Go types that mirror the proto definitions
// Will be replaced by proto-generated types after buf configuration is fixed

// QueryFeeStatisticsRequest is the request type for the Query/FeeStatistics RPC method
type QueryFeeStatisticsRequest struct{}

// QueryFeeStatisticsResponse is the response type for the Query/FeeStatistics RPC method
type QueryFeeStatisticsResponse struct {
	TotalCollectedAllTime    math.Int `json:"total_collected_all_time"`
	TotalBurnedAllTime       math.Int `json:"total_burned_all_time"`
	TotalToValidatorsAllTime math.Int `json:"total_to_validators_all_time"`
	TotalToTreasuryAllTime   math.Int `json:"total_to_treasury_all_time"`
	TotalTransactionsAllTime uint64   `json:"total_transactions_all_time"`
	LastUpdateHeight         int64    `json:"last_update_height"`
	CurrentEpoch             int64    `json:"current_epoch"`
}

// QueryBurnStatisticsRequest is the request type for the Query/BurnStatistics RPC method
type QueryBurnStatisticsRequest struct{}

// QueryBurnStatisticsResponse is the response type for the Query/BurnStatistics RPC method
type QueryBurnStatisticsResponse struct {
	TotalBurned     math.Int `json:"total_burned"`
	BurnRatePerDay  math.Int `json:"burn_rate_per_day"`
	CurrentSupply   math.Int `json:"current_supply"`
	BurnCapSupply   math.Int `json:"burn_cap_supply"`
	RemainingToCap  math.Int `json:"remaining_to_cap"`
	BurnCapReached  bool     `json:"burn_cap_reached"`
	LastBurnHeight  int64    `json:"last_burn_height"`
}

// QuerySupplySnapshotRequest is the request type for the Query/SupplySnapshot RPC method
type QuerySupplySnapshotRequest struct {
	Height int64 `json:"height"`
}

// QuerySupplySnapshotResponse is the response type for the Query/SupplySnapshot RPC method
type QuerySupplySnapshotResponse struct {
	Height            int64    `json:"height"`
	Timestamp         int64    `json:"timestamp"`
	TotalSupply       math.Int `json:"total_supply"`
	CirculatingSupply math.Int `json:"circulating_supply"`
	LiquidSupply      math.Int `json:"liquid_supply"`
	BondedSupply      math.Int `json:"bonded_supply"`
	BurnedCumulative  math.Int `json:"burned_cumulative"`
}

// QuerySupplySnapshotLatestRequest is the request type for the Query/SupplySnapshotLatest RPC method
type QuerySupplySnapshotLatestRequest struct{}

// QuerySupplySnapshotLatestResponse is the response type for the Query/SupplySnapshotLatest RPC method
type QuerySupplySnapshotLatestResponse struct {
	Height            int64    `json:"height"`
	Timestamp         int64    `json:"timestamp"`
	TotalSupply       math.Int `json:"total_supply"`
	CirculatingSupply math.Int `json:"circulating_supply"`
	LiquidSupply      math.Int `json:"liquid_supply"`
	BondedSupply      math.Int `json:"bonded_supply"`
	BurnedCumulative  math.Int `json:"burned_cumulative"`
}

// QueryFeeAccumulatorRequest is the request type for the Query/FeeAccumulator RPC method
type QueryFeeAccumulatorRequest struct{}

// QueryFeeAccumulatorResponse is the response type for the Query/FeeAccumulator RPC method
type QueryFeeAccumulatorResponse struct {
	Height           int64    `json:"height"`
	TotalCollected   math.Int `json:"total_collected"`
	TransactionCount uint64   `json:"transaction_count"`
}

// Implement proto.Message interface for compatibility (minimal implementation)

func (m *QueryFeeStatisticsRequest) Reset()         {}
func (m *QueryFeeStatisticsRequest) String() string { return "QueryFeeStatisticsRequest{}" }
func (m *QueryFeeStatisticsRequest) ProtoMessage()  {}

func (m *QueryFeeStatisticsResponse) Reset()         {}
func (m *QueryFeeStatisticsResponse) String() string { return "QueryFeeStatisticsResponse{}" }
func (m *QueryFeeStatisticsResponse) ProtoMessage()  {}

func (m *QueryBurnStatisticsRequest) Reset()         {}
func (m *QueryBurnStatisticsRequest) String() string { return "QueryBurnStatisticsRequest{}" }
func (m *QueryBurnStatisticsRequest) ProtoMessage()  {}

func (m *QueryBurnStatisticsResponse) Reset()         {}
func (m *QueryBurnStatisticsResponse) String() string { return "QueryBurnStatisticsResponse{}" }
func (m *QueryBurnStatisticsResponse) ProtoMessage()  {}

func (m *QuerySupplySnapshotRequest) Reset()         {}
func (m *QuerySupplySnapshotRequest) String() string { return "QuerySupplySnapshotRequest{}" }
func (m *QuerySupplySnapshotRequest) ProtoMessage()  {}

func (m *QuerySupplySnapshotResponse) Reset()         {}
func (m *QuerySupplySnapshotResponse) String() string { return "QuerySupplySnapshotResponse{}" }
func (m *QuerySupplySnapshotResponse) ProtoMessage()  {}

func (m *QuerySupplySnapshotLatestRequest) Reset()         {}
func (m *QuerySupplySnapshotLatestRequest) String() string { return "QuerySupplySnapshotLatestRequest{}" }
func (m *QuerySupplySnapshotLatestRequest) ProtoMessage()  {}

func (m *QuerySupplySnapshotLatestResponse) Reset()         {}
func (m *QuerySupplySnapshotLatestResponse) String() string { return "QuerySupplySnapshotLatestResponse{}" }
func (m *QuerySupplySnapshotLatestResponse) ProtoMessage()  {}

func (m *QueryFeeAccumulatorRequest) Reset()         {}
func (m *QueryFeeAccumulatorRequest) String() string { return "QueryFeeAccumulatorRequest{}" }
func (m *QueryFeeAccumulatorRequest) ProtoMessage()  {}

func (m *QueryFeeAccumulatorResponse) Reset()         {}
func (m *QueryFeeAccumulatorResponse) String() string { return "QueryFeeAccumulatorResponse{}" }
func (m *QueryFeeAccumulatorResponse) ProtoMessage()  {}
