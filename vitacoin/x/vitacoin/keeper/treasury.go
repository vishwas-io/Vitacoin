package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ===========================================================================================
// PHASE 3 - TASK 3.4: TREASURY MODULE & GOVERNANCE INTEGRATION
// ===========================================================================================
//
// This file implements the production-grade treasury management system for VITACOIN.
// The treasury collects 25% of protocol fees and is controlled exclusively by governance.
//
// KEY FEATURES:
// - Governance-controlled spending via proposals
// - Complete audit trail of all treasury operations
// - Balance queries and historical tracking
// - Integration with x/gov for proposal-based spending
// - Security validations and authorization checks
//
// ARCHITECTURE:
// - Treasury funds stored in dedicated module account (vitacoin_treasury)
// - All spending requires governance approval
// - Historical records maintained for transparency
// - Events emitted for all treasury operations
//
// ===========================================================================================

// --- Treasury Module Account Management ---

// GetTreasuryModuleAccount retrieves the treasury module account with validation
// The treasury account holds protocol fees designated for ecosystem development
func (k Keeper) GetTreasuryModuleAccount(ctx context.Context) (sdk.ModuleAccountI, error) {
	treasuryAcc := k.accountKeeper.GetModuleAccount(ctx, types.TreasuryModuleName)
	if treasuryAcc == nil {
		return nil, fmt.Errorf("treasury module account not found: %s", types.TreasuryModuleName)
	}
	
	k.logger.Debug("Retrieved treasury module account",
		"module", types.TreasuryModuleName,
		"address", treasuryAcc.GetAddress().String(),
	)
	
	return treasuryAcc, nil
}

// GetVitacoinModuleAccount retrieves the main vitacoin module account
// This account handles fee collection and escrow operations
func (k Keeper) GetVitacoinModuleAccount(ctx context.Context) (sdk.ModuleAccountI, error) {
	vitacoinAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if vitacoinAcc == nil {
		return nil, fmt.Errorf("vitacoin module account not found: %s", types.ModuleName)
	}
	
	k.logger.Debug("Retrieved vitacoin module account",
		"module", types.ModuleName,
		"address", vitacoinAcc.GetAddress().String(),
	)
	
	return vitacoinAcc, nil
}

// --- Treasury Balance Queries ---

// GetTreasuryBalance returns the current balance of the treasury module account
// Returns balance in native VITA tokens
func (k Keeper) GetTreasuryBalance(ctx context.Context) (sdk.Coins, error) {
	treasuryAcc, err := k.GetTreasuryModuleAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get treasury module account: %w", err)
	}
	
	balance := k.bankKeeper.GetAllBalances(ctx, treasuryAcc.GetAddress())
	
	k.logger.Debug("Retrieved treasury balance",
		"balance", balance.String(),
	)
	
	return balance, nil
}

// GetTreasuryBalanceDenom returns the treasury balance for a specific denomination
func (k Keeper) GetTreasuryBalanceDenom(ctx context.Context, denom string) (sdk.Coin, error) {
	treasuryAcc, err := k.GetTreasuryModuleAccount(ctx)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to get treasury module account: %w", err)
	}
	
	balance := k.bankKeeper.GetBalance(ctx, treasuryAcc.GetAddress(), denom)
	
	k.logger.Debug("Retrieved treasury balance for denom",
		"denom", denom,
		"balance", balance.String(),
	)
	
	return balance, nil
}

// GetVitaTreasuryBalance returns the VITA token balance in the treasury
// Convenience method for the primary token denomination
func (k Keeper) GetVitaTreasuryBalance(ctx context.Context) (math.Int, error) {
	// Use native denom (avita) - BondDenom not in params yet
	balance, err := k.GetTreasuryBalanceDenom(ctx, "avita")
	if err != nil {
		return math.ZeroInt(), err
	}
	
	return balance.Amount, nil
}

// --- Treasury Spending Operations ---

// SpendFromTreasury executes a treasury spending operation
// This function should ONLY be called from a governance proposal handler
// 
// Authorization Flow:
// 1. Governance proposal created with TreasurySpendProposal
// 2. Proposal passes voting period
// 3. Handler calls this function with validated parameters
// 4. Funds transferred to recipient
// 5. Spending record created for audit trail
func (k Keeper) SpendFromTreasury(
	ctx context.Context,
	recipient string,
	amount sdk.Coins,
	purpose string,
	proposalID uint64,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Validate recipient address
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address %s: %w", recipient, err)
	}
	
	// Validate amount
	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("invalid amount: %s", amount)
	}
	
	// Validate purpose is not empty
	if purpose == "" {
		return fmt.Errorf("spending purpose cannot be empty")
	}
	
	// Check treasury has sufficient balance
	if err := k.ValidateTreasurySpending(ctx, amount); err != nil {
		return fmt.Errorf("treasury spending validation failed: %w", err)
	}
	
	// Transfer funds from treasury to recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.TreasuryModuleName,
		recipientAddr,
		amount,
	); err != nil {
		return fmt.Errorf("failed to send coins from treasury: %w", err)
	}
	
	// Create spending record for audit trail
	spendingRecord := types.TreasurySpending{
		Id:          fmt.Sprintf("treasury-spend-%d-%d", proposalID, sdkCtx.BlockHeight()),
		ProposalId:  proposalID,
		Recipient:   recipient,
		Amount:      amount,
		Purpose:     purpose,
		SpentHeight: sdkCtx.BlockHeight(),
		SpentTime:   sdkCtx.BlockTime().Unix(),
	}
	
	if err := k.SetTreasurySpending(ctx, spendingRecord); err != nil {
		// Don't fail the spending if record storage fails, just log error
		k.logger.Error("failed to store treasury spending record",
			"error", err,
			"spending_id", spendingRecord.Id,
		)
	}
	
	// Emit event for treasury spending
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasurySpent,
			sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyPurpose, purpose),
			sdk.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
		),
	)
	
	k.logger.Info("Treasury spending executed",
		"proposal_id", proposalID,
		"recipient", recipient,
		"amount", amount.String(),
		"purpose", purpose,
		"height", sdkCtx.BlockHeight(),
	)
	
	return nil
}

// ValidateTreasurySpending validates that treasury has sufficient funds for spending
// Checks balance against requested amount with safety margin
func (k Keeper) ValidateTreasurySpending(ctx context.Context, amount sdk.Coins) error {
	treasuryBalance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get treasury balance: %w", err)
	}
	
	// Check if treasury has sufficient balance
	if !treasuryBalance.IsAllGTE(amount) {
		return fmt.Errorf("insufficient treasury balance: have %s, need %s", 
			treasuryBalance, amount)
	}
	
	// Additional validation: ensure we're not spending the entire treasury
	// Keep at least 1% as buffer for safety
	vitaBalance := treasuryBalance.AmountOf("avita")
	requestedVita := amount.AmountOf("avita")
	
	if !requestedVita.IsZero() {
		// Calculate 99% of balance
		ninetyNinePercent := vitaBalance.Mul(math.NewInt(99)).Quo(math.NewInt(100))
		
		if requestedVita.GT(ninetyNinePercent) {
			return fmt.Errorf("spending amount exceeds 99%% of treasury balance: requested %s, max allowed %s", 
				requestedVita, ninetyNinePercent)
		}
	}
	
	k.logger.Debug("Treasury spending validation passed",
		"amount", amount.String(),
		"treasury_balance", treasuryBalance.String(),
	)
	
	return nil
}

// --- Treasury Spending History ---

// SetTreasurySpending stores a treasury spending record
func (k Keeper) SetTreasurySpending(ctx context.Context, spending types.TreasurySpending) error {
	if spending.Id == "" {
		return fmt.Errorf("treasury spending ID cannot be empty")
	}
	
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&spending)
	
	return store.Set(types.GetTreasurySpendingKey(spending.Id), bz)
}

// GetTreasurySpending retrieves a treasury spending record by ID
func (k Keeper) GetTreasurySpending(ctx context.Context, id string) (types.TreasurySpending, error) {
	store := k.storeService.OpenKVStore(ctx)
	
	bz, err := store.Get(types.GetTreasurySpendingKey(id))
	if err != nil {
		return types.TreasurySpending{}, err
	}
	
	if bz == nil {
		return types.TreasurySpending{}, fmt.Errorf("treasury spending not found: %s", id)
	}
	
	var spending types.TreasurySpending
	k.cdc.MustUnmarshal(bz, &spending)
	return spending, nil
}

// GetAllTreasurySpending retrieves all treasury spending records
func (k Keeper) GetAllTreasurySpending(ctx context.Context) ([]types.TreasurySpending, error) {
	store := k.storeService.OpenKVStore(ctx)
	spending := []types.TreasurySpending{}
	
	iter, err := store.Iterator(types.TreasurySpendingKeyPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	
	for ; iter.Valid(); iter.Next() {
		var record types.TreasurySpending
		k.cdc.MustUnmarshal(iter.Value(), &record)
		spending = append(spending, record)
	}
	
	return spending, nil
}

// GetTreasurySpendingByProposal retrieves all spending records for a specific proposal
func (k Keeper) GetTreasurySpendingByProposal(ctx context.Context, proposalID uint64) ([]types.TreasurySpending, error) {
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, err
	}
	
	filtered := []types.TreasurySpending{}
	for _, spending := range allSpending {
		if spending.ProposalId == proposalID {
			filtered = append(filtered, spending)
		}
	}
	
	return filtered, nil
}

// GetTreasurySpendingByRecipient retrieves all spending records for a specific recipient
func (k Keeper) GetTreasurySpendingByRecipient(ctx context.Context, recipient string) ([]types.TreasurySpending, error) {
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, err
	}
	
	filtered := []types.TreasurySpending{}
	for _, spending := range allSpending {
		if spending.Recipient == recipient {
			filtered = append(filtered, spending)
		}
	}
	
	return filtered, nil
}

// GetTreasurySpendingInRange retrieves spending records within a height range
func (k Keeper) GetTreasurySpendingInRange(ctx context.Context, startHeight, endHeight int64) ([]types.TreasurySpending, error) {
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, err
	}
	
	filtered := []types.TreasurySpending{}
	for _, spending := range allSpending {
		if spending.SpentHeight >= startHeight && spending.SpentHeight <= endHeight {
			filtered = append(filtered, spending)
		}
	}
	
	return filtered, nil
}

// --- Treasury Statistics ---

// GetTreasuryStatistics calculates comprehensive treasury statistics
func (k Keeper) GetTreasuryStatistics(ctx context.Context) (*types.TreasuryStatistics, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Get current balance
	balance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get treasury balance: %w", err)
	}
	
	// Get all spending records
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending records: %w", err)
	}
	
	// Calculate total spent
	totalSpent := sdk.NewCoins()
	for _, spending := range allSpending {
		totalSpent = totalSpent.Add(spending.Amount...)
	}
	
	// Get fee statistics to calculate total treasury deposits
	feeStats, err := k.GetFeeStatistics(ctx)
	if err != nil {
		// If fee stats don't exist yet, create empty stats
		feeStats = types.FeeStatistics{
			TotalCollectedAllTime:    math.ZeroInt(),
			TotalBurnedAllTime:       math.ZeroInt(),
			TotalToValidatorsAllTime: math.ZeroInt(),
			TotalToTreasuryAllTime:   math.ZeroInt(),
			TotalTransactionsAllTime: 0,
			LastUpdateHeight:         sdkCtx.BlockHeight(),
			CurrentEpoch:             k.CalculateEpoch(sdkCtx),
		}
	}
	
	stats := &types.TreasuryStatistics{
		CurrentBalance:     balance,
		TotalDeposited:     sdk.NewCoins(sdk.NewCoin("avita", feeStats.TotalToTreasuryAllTime)),
		TotalSpent:         totalSpent,
		SpendingCount:      uint64(len(allSpending)),
		LastUpdateHeight:   sdkCtx.BlockHeight(),
		LastUpdateTime:     sdkCtx.BlockTime().Unix(),
	}
	
	k.logger.Debug("Calculated treasury statistics",
		"current_balance", stats.CurrentBalance.String(),
		"total_deposited", stats.TotalDeposited.String(),
		"total_spent", stats.TotalSpent.String(),
		"spending_count", stats.SpendingCount,
	)
	
	return stats, nil
}

// --- Treasury Integration with Fee Distribution ---

// DepositToTreasury deposits funds to the treasury module account
// This is called by the fee distribution mechanism
func (k Keeper) DepositToTreasury(ctx context.Context, amount sdk.Coins) error {
	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("invalid deposit amount: %s", amount)
	}
	
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Transfer from vitacoin module to treasury module
	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		types.TreasuryModuleName,
		amount,
	); err != nil {
		return fmt.Errorf("failed to deposit to treasury: %w", err)
	}
	
	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryDeposit,
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
		),
	)
	
	k.logger.Info("Deposited to treasury",
		"amount", amount.String(),
		"height", sdkCtx.BlockHeight(),
	)
	
	return nil
}

// --- Helper Functions ---

// FormatTreasuryBalance formats treasury balance for display
func (k Keeper) FormatTreasuryBalance(ctx context.Context) (string, error) {
	balance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return "", err
	}
	
	if balance.IsZero() {
		return "0 VITA", nil
	}
	
	return balance.String(), nil
}

// GetTreasuryAgeInBlocks returns how many blocks the treasury has been active
func (k Keeper) GetTreasuryAgeInBlocks(ctx context.Context) int64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	
	// Get the first spending record to estimate treasury start
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil || len(allSpending) == 0 {
		return 0
	}
	
	// Find earliest spending
	earliestHeight := sdkCtx.BlockHeight()
	for _, spending := range allSpending {
		if spending.SpentHeight < earliestHeight {
			earliestHeight = spending.SpentHeight
		}
	}
	
	return sdkCtx.BlockHeight() - earliestHeight
}

// EstimateTreasuryRunway estimates how long treasury can sustain current spending rate
// Returns estimated blocks until treasury depletion (0 if no spending history)
func (k Keeper) EstimateTreasuryRunway(ctx context.Context) (int64, error) {
	balance, err := k.GetVitaTreasuryBalance(ctx)
	if err != nil {
		return 0, err
	}
	
	if balance.IsZero() {
		return 0, nil
	}
	
	// Get spending history
	allSpending, err := k.GetAllTreasurySpending(ctx)
	if err != nil || len(allSpending) == 0 {
		return 0, nil // No spending history, can't estimate
	}
	
	// Calculate total spent amount (in avita)
	totalSpent := math.ZeroInt()
	for _, spending := range allSpending {
		totalSpent = totalSpent.Add(spending.Amount.AmountOf("avita"))
	}
	
	if totalSpent.IsZero() {
		return 0, nil // No spending yet
	}
	
	// Calculate age in blocks
	age := k.GetTreasuryAgeInBlocks(ctx)
	if age == 0 {
		return 0, nil
	}
	
	// Calculate average spending per block
	avgSpendingPerBlock := totalSpent.Quo(math.NewInt(age))
	
	if avgSpendingPerBlock.IsZero() {
		return 0, nil // Negligible spending
	}
	
	// Estimate blocks until depletion
	runway := balance.Quo(avgSpendingPerBlock)
	
	k.logger.Debug("Estimated treasury runway",
		"balance", balance.String(),
		"avg_spending_per_block", avgSpendingPerBlock.String(),
		"estimated_blocks", runway.Int64(),
		"estimated_days", runway.Int64()/14400, // ~14400 blocks per day
	)
	
	return runway.Int64(), nil
}

// GetTreasuryHealth returns a health score for the treasury (0-100)
// 100 = Healthy (growing balance), 0 = Critical (depleting rapidly)
func (k Keeper) GetTreasuryHealth(ctx context.Context) (uint32, error) {
	balance, err := k.GetVitaTreasuryBalance(ctx)
	if err != nil {
		return 0, err
	}
	
	if balance.IsZero() {
		return 0, nil // Empty treasury = 0 health
	}
	
	runway, err := k.EstimateTreasuryRunway(ctx)
	if err != nil {
		return 50, nil // Unknown runway, return neutral score
	}
	
	if runway == 0 {
		return 100, nil // No spending or infinite runway
	}
	
	// Health scoring:
	// > 1 year runway = 100
	// 6 months runway = 75
	// 3 months runway = 50
	// 1 month runway = 25
	// < 1 month runway = 10
	
	oneYear := int64(5256000)   // ~365 days of blocks
	sixMonths := oneYear / 2
	threeMonths := oneYear / 4
	oneMonth := oneYear / 12
	
	if runway >= oneYear {
		return 100, nil
	} else if runway >= sixMonths {
		return 75, nil
	} else if runway >= threeMonths {
		return 50, nil
	} else if runway >= oneMonth {
		return 25, nil
	}
	
	return 10, nil // Critical
}

// --- Genesis Export ---

// ExportTreasuryGenesis exports treasury state for genesis
func (k Keeper) ExportTreasuryGenesis(ctx context.Context) (*types.TreasuryGenesisState, error) {
	balance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get treasury balance: %w", err)
	}
	
	spending, err := k.GetAllTreasurySpending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get treasury spending: %w", err)
	}
	
	return &types.TreasuryGenesisState{
		Balance:      balance,
		SpendingList: spending,
	}, nil
}

// ImportTreasuryGenesis imports treasury state from genesis
func (k Keeper) ImportTreasuryGenesis(ctx context.Context, genesis *types.TreasuryGenesisState) error {
	// Import spending records
	for _, spending := range genesis.SpendingList {
		if err := k.SetTreasurySpending(ctx, spending); err != nil {
			return fmt.Errorf("failed to import treasury spending %s: %w", spending.Id, err)
		}
	}
	
	k.logger.Info("Imported treasury genesis state",
		"spending_records", len(genesis.SpendingList),
		"balance", genesis.Balance.String(),
	)
	
	return nil
}
