package keeper

import (
	"context"
	"fmt"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	
	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ===========================================================================================
// PHASE 3 - TASK 3.4: TREASURY GOVERNANCE PROPOSAL HANDLER
// ===========================================================================================
//
// Implements governance proposal handler for treasury spending operations
//
// SECURITY:
// - Only executable through governance
// - Requires community voting approval
// - Full validation before execution
// - Complete audit trail maintained
//
// ===========================================================================================

// HandleTreasurySpendProposal handles a treasury spend proposal after it passes governance
// This is the ONLY way to spend treasury funds - governance approval is mandatory
func HandleTreasurySpendProposal(ctx sdk.Context, k Keeper, proposal *types.TreasurySpendProposal) error {
	// Validate proposal
	if err := proposal.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid treasury spend proposal: %w", err)
	}
	
	k.logger.Info("Processing treasury spend proposal",
		"title", proposal.Title,
		"recipient", proposal.Recipient,
		"amount", proposal.Amount.String(),
	)
	
	// Get next proposal ID (this would come from gov keeper in production)
	// For now, use block height as pseudo-proposal-id
	proposalID := uint64(ctx.BlockHeight())
	
	// Execute the spending
	if err := k.SpendFromTreasury(
		ctx,
		proposal.Recipient,
		proposal.Amount,
		proposal.Description,
		proposalID,
	); err != nil {
		return fmt.Errorf("failed to execute treasury spending: %w", err)
	}
	
	k.logger.Info("Treasury spend proposal executed successfully",
		"title", proposal.Title,
		"recipient", proposal.Recipient,
		"amount", proposal.Amount.String(),
		"proposal_id", proposalID,
	)
	
	return nil
}

// NewTreasurySpendProposalHandler creates a governance proposal handler for treasury spending
// This integrates with x/gov module for governance-controlled treasury operations
// TODO: Enable when TreasurySpendProposal implements govtypes.Content interface (after proto regen)
func NewTreasurySpendProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		// TODO: Uncomment after proto regeneration
		// switch c := content.(type) {
		// case *types.TreasurySpendProposal:
		// 	return HandleTreasurySpendProposal(ctx, k, c)
		// default:
		// 	return fmt.Errorf("unrecognized treasury proposal content type: %T", c)
		// }
		return fmt.Errorf("treasury spend proposal handler not yet integrated (pending proto regen)")
	}
}

// ValidateTreasurySpendProposal performs comprehensive validation of a treasury spend proposal
// This should be called before proposal submission
func (k Keeper) ValidateTreasurySpendProposal(ctx context.Context, proposal *types.TreasurySpendProposal) error {
	// Basic validation
	if err := proposal.ValidateBasic(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}
	
	// Validate recipient is not a module account (except treasury itself for special operations)
	recipientAddr, _ := sdk.AccAddressFromBech32(proposal.Recipient)
	account := k.accountKeeper.GetAccount(ctx, recipientAddr)
	if account != nil {
		if _, isModuleAcc := account.(sdk.ModuleAccountI); isModuleAcc {
			// Only allow spending to treasury itself (for rollovers, etc)
			if proposal.Recipient != k.accountKeeper.GetModuleAddress(types.TreasuryModuleName).String() {
				return fmt.Errorf("cannot spend treasury funds to module accounts other than treasury")
			}
		}
	}
	
	// Validate treasury has sufficient balance
	if err := k.ValidateTreasurySpending(ctx, proposal.Amount); err != nil {
		return fmt.Errorf("insufficient treasury funds: %w", err)
	}
	
	// Validate amount is reasonable (not dust, not excessive)
	vitaAmount := proposal.Amount.AmountOf("avita")
	if !vitaAmount.IsZero() {
		// Minimum spend: 1 VITA (prevents spam proposals)
		minSpend := math.NewInt(1_000_000_000_000_000_000) // 1 VITA in avita
		if vitaAmount.LT(minSpend) {
			return fmt.Errorf("treasury spend amount too small: minimum %s, got %s", 
				minSpend, vitaAmount)
		}
	}
	
	k.logger.Debug("Treasury spend proposal validation passed",
		"recipient", proposal.Recipient,
		"amount", proposal.Amount.String(),
	)
	
	return nil
}

// EstimateTreasurySpendImpact calculates the impact of a proposed spending on treasury health
// Returns:
// - new balance after spend
// - new runway in blocks
// - new health score
// - whether spending is recommended
func (k Keeper) EstimateTreasurySpendImpact(
	ctx context.Context,
	amount sdk.Coins,
) (*types.TreasuryImpactEstimate, error) {
	// Get current state
	currentBalance, err := k.GetTreasuryBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get treasury balance: %w", err)
	}
	
	currentRunway, err := k.EstimateTreasuryRunway(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate runway: %w", err)
	}
	
	currentHealth, err := k.GetTreasuryHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get health: %w", err)
	}
	
	// Calculate new balance
	newBalance := currentBalance.Sub(amount...)
	if newBalance.IsAnyNegative() {
		return nil, fmt.Errorf("spending would result in negative balance")
	}
	
	// Estimate new runway (simplified - assumes same spending rate)
	vitaBalance := newBalance.AmountOf("avita")
	currentVitaBalance := currentBalance.AmountOf("avita")
	
	var newRunway int64
	if currentRunway > 0 && !currentVitaBalance.IsZero() {
		// Scale runway proportionally to balance change
		newRunway = currentRunway * vitaBalance.Quo(currentVitaBalance).Int64()
	}
	
	// Estimate new health score
	var newHealth uint32
	oneYear := int64(5256000) // ~1 year in blocks
	if newRunway == 0 {
		newHealth = 100 // No spending history
	} else {
		sixMonths := oneYear / 2
		threeMonths := oneYear / 4
		oneMonth := oneYear / 12
		
		if newRunway >= oneYear {
			newHealth = 100
		} else if newRunway >= sixMonths {
			newHealth = 75
		} else if newRunway >= threeMonths {
			newHealth = 50
		} else if newRunway >= oneMonth {
			newHealth = 25
		} else {
			newHealth = 10
		}
	}
	
	// Determine if spending is recommended
	// Don't recommend if it drops health below 25 or runway below 1 month
	recommended := newHealth >= 25 && (newRunway == 0 || newRunway >= oneYear/12)
	
	estimate := &types.TreasuryImpactEstimate{
		CurrentBalance:  currentBalance,
		SpendAmount:     amount,
		NewBalance:      newBalance,
		CurrentRunway:   currentRunway,
		NewRunway:       newRunway,
		CurrentHealth:   currentHealth,
		NewHealth:       newHealth,
		Recommended:     recommended,
	}
	
	k.logger.Debug("Estimated treasury spend impact",
		"current_balance", currentBalance.String(),
		"spend_amount", amount.String(),
		"new_balance", newBalance.String(),
		"current_health", currentHealth,
		"new_health", newHealth,
		"recommended", recommended,
	)
	
	return estimate, nil
}

// GetTreasurySpendingReport generates a comprehensive report of treasury activity
// Useful for governance proposals and community transparency
func (k Keeper) GetTreasurySpendingReport(ctx context.Context, fromHeight, toHeight int64) (*types.TreasurySpendingReport, error) {
	// Get spending in range
	spending, err := k.GetTreasurySpendingInRange(ctx, fromHeight, toHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending: %w", err)
	}
	
	// Calculate totals
	totalSpent := sdk.NewCoins()
	recipientMap := make(map[string]sdk.Coins)
	purposeMap := make(map[string]sdk.Coins)
	
	for _, record := range spending {
		totalSpent = totalSpent.Add(record.Amount...)
		
		// Aggregate by recipient
		if existing, ok := recipientMap[record.Recipient]; ok {
			recipientMap[record.Recipient] = existing.Add(record.Amount...)
		} else {
			recipientMap[record.Recipient] = record.Amount
		}
		
		// Aggregate by purpose
		if existing, ok := purposeMap[record.Purpose]; ok {
			purposeMap[record.Purpose] = existing.Add(record.Amount...)
		} else {
			purposeMap[record.Purpose] = record.Amount
		}
	}
	
	// Get current stats
	stats, err := k.GetTreasuryStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	
	report := &types.TreasurySpendingReport{
		FromHeight:      fromHeight,
		ToHeight:        toHeight,
		TotalSpent:      totalSpent,
		SpendingCount:   uint64(len(spending)),
		SpendingRecords: spending,
		ByRecipient:     recipientMap,
		ByPurpose:       purposeMap,
		CurrentStats:    stats,
	}
	
	k.logger.Info("Generated treasury spending report",
		"from_height", fromHeight,
		"to_height", toHeight,
		"total_spent", totalSpent.String(),
		"spending_count", len(spending),
	)
	
	return report, nil
}
