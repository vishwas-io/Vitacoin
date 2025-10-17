package types

import (
	"cosmossdk.io/math"
	"time"
)

// Phase 3: Fee and Treasury Types
// These are simplified Go structs that will be replaced by proto-generated types later

// BlockFeeAccumulator represents temporary accumulator for current block's fees
type BlockFeeAccumulator struct {
	Height           int64
	TotalCollected   math.Int
	TransactionCount uint64
}

// FeeStatistics tracks cumulative fee statistics since genesis
type FeeStatistics struct {
	TotalCollectedAllTime    math.Int
	TotalBurnedAllTime       math.Int
	TotalToValidatorsAllTime math.Int
	TotalToTreasuryAllTime   math.Int
	TotalTransactionsAllTime uint64
	LastUpdateHeight         int64
	CurrentEpoch             int64
}

// BurnStats tracks burn mechanism statistics
type BurnStats struct {
	TotalBurned     math.Int
	BurnRatePerDay  math.Int
	CurrentSupply   math.Int
	BurnCapSupply   math.Int
	RemainingToCap  math.Int
	BurnCapReached  bool
	LastBurnHeight  int64
}

// SupplySnapshot tracks supply metrics at a specific point in time
type SupplySnapshot struct {
	Height            int64
	Timestamp         time.Time
	TotalSupply       math.Int
	CirculatingSupply math.Int
	LiquidSupply      math.Int
	BondedSupply      math.Int
	BurnedCumulative  math.Int
}

// Params - extend existing Params struct (this will be added to params.pb.go after proto regen)
// For now, we'll add these fields manually to the DefaultParams and Validate functions

// Phase 3 additional param fields (to be added to Params proto):
// - FeeValidatorPercent    math.LegacyDec
// - FeeTreasuryPercent     math.LegacyDec
// - MinProtocolFee         math.Int
// - MaxProtocolFee         math.Int
// - BurnCapSupply          math.Int
// - PausedFeeCollection    bool
// - PausedFeeDistribution  bool
