package types

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Advanced validation constants
const (
	// Business name constraints
	MinBusinessNameLength = 3
	MaxBusinessNameLength = 100
	
	// Payment constraints — uvita (6 decimals: 1 VITA = 1e6 uvita)
	MinPaymentAmount     = 1000        // 0.001 VITA (1e3 uvita)
	MaxPaymentAmount     = 1000000000000 // 1,000,000 VITA (1e12 uvita)
	MaxMemoLength        = 256
	
	// Vault constraints — uvita (6 decimals)
	MinVaultAmount       = 1000000     // 1 VITA (1e6 uvita)
	MaxVaultAmount       = 10000000000000 // 10,000,000 VITA (1e13 uvita)
	MinLockDuration      = 1     // At least 1 block
	MaxLockDuration      = 5_256_000 // ~1 year at 6s/block (aligned with tests)
	MaxUnlockHeight      = 5_256_000 // Maximum unlock height (~1 year)
	
	// Pool constraints — uvita (6 decimals)
	MinPoolNameLength    = 3
	MaxPoolNameLength    = 50
	MaxPoolDuration      = 100_000_000 // 100M blocks (aligned with tests)
	MinPoolAmount        = 1000        // 0.001 VITA (1e3 uvita)
	
	// Merchant tier thresholds — uvita (6 decimals)
	TierBronzeThreshold  = 10000000000   // 10,000 VITA (1e10 uvita)
	TierSilverThreshold  = 50000000000   // 50,000 VITA (5e10 uvita)
	TierGoldThreshold    = 100000000000  // 100,000 VITA (1e11 uvita)
	
	// Security constraints
	MaxRecipientsPerDistribution = 1000 // Prevent spam
	MaxStringLength              = 1000 // General string length limit
)

// Regular expressions for validation
var (
	// Allow alphanumeric, spaces, and common business characters
	businessNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,&()]+$`)
	
	// Pool names should be alphanumeric with spaces and hyphens
	poolNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	
	// Memo can contain most printable characters but no control characters
	memoRegex = regexp.MustCompile(`^[\x20-\x7E\t\n\r]*$`)
)

// Enhanced validation functions

// ValidateBusinessName performs comprehensive business name validation
func ValidateBusinessName(name string) error {
	if name == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("business name cannot be empty")
	}
	
	if len(name) < MinBusinessNameLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("business name must be at least %d characters", MinBusinessNameLength)
	}
	
	if len(name) > MaxBusinessNameLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("business name cannot exceed %d characters", MaxBusinessNameLength)
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("business name contains invalid UTF-8 characters")
	}
	
	// Check for control characters (except allowed ones)
	if containsInvalidControlChars(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("business name contains invalid control characters")
	}
	
	// Check against regex pattern
	if !businessNameRegex.MatchString(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("business name contains invalid characters")
	}
	
	// Additional checks
	trimmed := strings.TrimSpace(name)
	if trimmed != name {
		return sdkerrors.ErrInvalidRequest.Wrap("business name cannot start or end with whitespace")
	}
	
	// Check for excessive whitespace
	if strings.Contains(name, "  ") {
		return sdkerrors.ErrInvalidRequest.Wrap("business name cannot contain consecutive spaces")
	}
	
	return nil
}

// ValidatePaymentAmount performs comprehensive payment amount validation
func ValidatePaymentAmount(amount math.Int) error {
	if amount.IsNil() {
		return sdkerrors.ErrInvalidRequest.Wrap("amount cannot be nil")
	}
	
	if amount.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("amount cannot be negative")
	}
	
	if amount.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("amount must be positive")
	}
	
	// Check minimum amount (prevent dust attacks)
	if amount.LT(math.NewInt(MinPaymentAmount)) {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount must be at least %s uvita", math.NewInt(MinPaymentAmount).String())
	}
	
	// Check maximum amount (prevent overflow attacks)
	// 1e24 = 1,000,000 VITA (need to construct from smaller numbers to avoid overflow)
	maxAmount := math.NewInt(MaxPaymentAmount) // 1M VITA in uvita
	if amount.GT(maxAmount) {
		return sdkerrors.ErrInvalidRequest.Wrap("amount exceeds maximum allowed")
	}
	
	return nil
}

// ValidateVaultAmount performs comprehensive vault amount validation
func ValidateVaultAmount(amount math.Int) error {
	if amount.IsNil() {
		return sdkerrors.ErrInvalidRequest.Wrap("vault amount cannot be nil")
	}
	
	if amount.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("vault amount cannot be negative")
	}
	
	if amount.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("vault amount must be positive")
	}
	
	// Check minimum vault amount
	if amount.LT(math.NewInt(MinVaultAmount)) {
		return sdkerrors.ErrInvalidRequest.Wrapf("vault amount must be at least %s uvita (1 VITA)", math.NewInt(MinVaultAmount).String())
	}
	
	// Check maximum vault amount
	// 1e25 = 10,000,000 VITA (need to construct from smaller numbers to avoid overflow)
	maxAmount := math.NewInt(MaxVaultAmount) // 10M VITA in uvita
	if amount.GT(maxAmount) {
		return sdkerrors.ErrInvalidRequest.Wrap("vault amount exceeds maximum allowed")
	}
	
	return nil
}

// ValidateLockDuration performs comprehensive lock duration validation
func ValidateLockDuration(duration uint64) error {
	if duration < MinLockDuration {
		return sdkerrors.ErrInvalidRequest.Wrapf("lock duration must be at least %d blocks", MinLockDuration)
	}
	
	if duration > MaxLockDuration {
		return sdkerrors.ErrInvalidRequest.Wrapf("lock duration cannot exceed %d blocks (~1 year)", MaxLockDuration)
	}
	
	return nil
}

// ValidateUnlockHeight performs unlock height validation
func ValidateUnlockHeight(unlockHeight int64, currentHeight int64) error {
	if unlockHeight <= currentHeight {
		return sdkerrors.ErrInvalidRequest.Wrap("unlock height must be in the future")
	}
	
	if unlockHeight > MaxUnlockHeight {
		return sdkerrors.ErrInvalidRequest.Wrapf("unlock height cannot exceed %d", MaxUnlockHeight)
	}
	
	// Check reasonable duration (not too far in future)
	duration := unlockHeight - currentHeight
	if duration > MaxLockDuration {
		return sdkerrors.ErrInvalidRequest.Wrapf("lock duration cannot exceed %d blocks (~1 year)", MaxLockDuration)
	}
	
	return nil
}

// ValidatePoolName performs comprehensive pool name validation
func ValidatePoolName(name string) error {
	if name == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("pool name cannot be empty")
	}
	
	if len(name) < MinPoolNameLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("pool name must be at least %d characters", MinPoolNameLength)
	}
	
	if len(name) > MaxPoolNameLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("pool name cannot exceed %d characters", MaxPoolNameLength)
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("pool name contains invalid UTF-8 characters")
	}
	
	// Check for control characters
	if containsInvalidControlChars(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("pool name contains invalid control characters")
	}
	
	// Check against regex pattern
	if !poolNameRegex.MatchString(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("pool name contains invalid characters")
	}
	
	// Additional checks
	trimmed := strings.TrimSpace(name)
	if trimmed != name {
		return sdkerrors.ErrInvalidRequest.Wrap("pool name cannot start or end with whitespace")
	}
	
	return nil
}

// ValidatePoolDuration performs pool duration validation
func ValidatePoolDuration(startHeight, endHeight int64) error {
	if endHeight == 0 {
		// Unlimited duration is allowed
		return nil
	}
	
	if endHeight <= startHeight {
		return sdkerrors.ErrInvalidRequest.Wrap("end height must be after start height")
	}
	
	duration := endHeight - startHeight
	if duration > MaxPoolDuration {
		return sdkerrors.ErrInvalidRequest.Wrapf("pool duration cannot exceed %d blocks (~2 years)", MaxPoolDuration)
	}
	
	return nil
}

// ValidateMemo performs comprehensive memo validation
func ValidateMemo(memo string) error {
	if len(memo) > MaxMemoLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("memo cannot exceed %d characters", MaxMemoLength)
	}
	
	// Allow empty memos
	if memo == "" {
		return nil
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(memo) {
		return sdkerrors.ErrInvalidRequest.Wrap("memo contains invalid UTF-8 characters")
	}
	
	// Check for invalid control characters (allow tab, newline, carriage return)
	if containsInvalidControlChars(memo) {
		return sdkerrors.ErrInvalidRequest.Wrap("memo contains invalid control characters")
	}
	
	return nil
}

// ValidateRewardDistribution performs comprehensive reward distribution validation
func ValidateRewardDistribution(recipients []string, amounts []math.Int) error {
	if len(recipients) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("recipients list cannot be empty")
	}
	
	if len(recipients) != len(amounts) {
		return sdkerrors.ErrInvalidRequest.Wrap("recipients and amounts lists must have the same length")
	}
	
	if len(recipients) > MaxRecipientsPerDistribution {
		return sdkerrors.ErrInvalidRequest.Wrapf("cannot distribute to more than %d recipients at once", MaxRecipientsPerDistribution)
	}
	
	// Track unique recipients to prevent duplicates
	recipientSet := make(map[string]bool)
	totalAmount := math.ZeroInt()
	
	for i, recipient := range recipients {
		// Validate recipient address
		if _, err := sdk.AccAddressFromBech32(recipient); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address at index %d: %s", i, err)
		}
		
		// Check for duplicate recipients
		if recipientSet[recipient] {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate recipient address: %s", recipient)
		}
		recipientSet[recipient] = true
		
		// Validate amount
		amount := amounts[i]
		if amount.IsNil() || amount.IsNegative() || amount.IsZero() {
			return sdkerrors.ErrInvalidRequest.Wrapf("amount at index %d must be positive", i)
		}
		
		// Prevent extremely small amounts (dust)
		if amount.LT(math.NewInt(MinPoolAmount)) {
			return sdkerrors.ErrInvalidRequest.Wrapf("amount at index %d is too small", i)
		}
		
		totalAmount = totalAmount.Add(amount)
	}
	
	// Check total distribution amount isn't excessive
	maxAmount := math.NewInt(1000000000000000) // 1B VITA in uvita (1e15)
	if totalAmount.GT(maxAmount) {
		return sdkerrors.ErrInvalidRequest.Wrap("total distribution amount exceeds maximum allowed")
	}
	
	return nil
}

// ValidateID performs comprehensive ID validation
func ValidateID(id, idType string) error {
	if id == "" {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s ID cannot be empty", idType)
	}
	
	if len(id) > MaxStringLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s ID cannot exceed %d characters", idType, MaxStringLength)
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(id) {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s ID contains invalid UTF-8 characters", idType)
	}
	
	// Check for control characters
	if containsInvalidControlChars(id) {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s ID contains invalid control characters", idType)
	}
	
	return nil
}

// ValidateReason performs reason validation (for refunds, etc.)
func ValidateReason(reason string) error {
	if reason == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("reason cannot be empty")
	}
	
	if len(reason) > MaxMemoLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("reason cannot exceed %d characters", MaxMemoLength)
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(reason) {
		return sdkerrors.ErrInvalidRequest.Wrap("reason contains invalid UTF-8 characters")
	}
	
	// Check for control characters
	if containsInvalidControlChars(reason) {
		return sdkerrors.ErrInvalidRequest.Wrap("reason contains invalid control characters")
	}
	
	return nil
}

// Helper functions

// containsInvalidControlChars checks for invalid control characters
// Allows tab (\t), newline (\n), and carriage return (\r)
func containsInvalidControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			// Allow tab, newline, carriage return
			if r != '\t' && r != '\n' && r != '\r' {
				return true
			}
		}
	}
	return false
}

// ValidateAddressPair ensures two addresses are different and valid
func ValidateAddressPair(addr1, addr2, name1, name2 string) error {
	// Validate first address
	if _, err := sdk.AccAddressFromBech32(addr1); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid %s address: %s", name1, err)
	}
	
	// Validate second address
	if _, err := sdk.AccAddressFromBech32(addr2); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid %s address: %s", name2, err)
	}
	
	// Ensure addresses are different
	if addr1 == addr2 {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s and %s cannot be the same address", name1, name2)
	}
	
	return nil
}

// Security validation functions

// ValidateNoReentrancy checks for potential reentrancy patterns in IDs
func ValidateNoReentrancy(id string) error {
	// Check for patterns that might indicate reentrancy attempts
	dangerousPatterns := []string{
		"../", "./", "\\", "<script", "javascript:", "data:", "vbscript:",
	}
	
	lowerID := strings.ToLower(id)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerID, pattern) {
			return sdkerrors.ErrInvalidRequest.Wrapf("ID contains potentially dangerous pattern: %s", pattern)
		}
	}
	
	return nil
}

// ValidateStringLength performs general string length validation
func ValidateStringLength(s, fieldName string, maxLen int) error {
	if len(s) > maxLen {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s cannot exceed %d characters", fieldName, maxLen)
	}
	return nil
}

// Rate limiting and DoS protection (for future implementation)

// ValidateTransactionFrequency validates transaction frequency to prevent spam
// This would be implemented with stateful checks in the keeper
func ValidateTransactionFrequency(senderAddr string, lastTxTime int64, currentTime int64) error {
	const minTimeBetweenTx = 1 // 1 second minimum between transactions
	
	if currentTime-lastTxTime < minTimeBetweenTx {
		return sdkerrors.ErrInvalidRequest.Wrap("transaction frequency too high, please wait")
	}
	
	return nil
}

// Business logic validation helpers

// ValidateStakeAmount validates stake amounts with tier considerations
func ValidateStakeAmount(amount math.Int, minRequired math.Int) error {
	if amount.IsNil() {
		return sdkerrors.ErrInvalidRequest.Wrap("stake amount cannot be nil")
	}
	
	if amount.IsNegative() || amount.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("stake amount must be positive")
	}
	
	if amount.LT(minRequired) {
		return sdkerrors.ErrInvalidRequest.Wrapf("amount is below minimum allowed")
	}
	
	// Check maximum stake amount (1e27 = 1 trillion VITA with 18 decimals)
	maxStake := math.NewInt(1000000000000000) // 1B VITA in uvita (1e15)
	if amount.GT(maxStake) {
		return sdkerrors.ErrInvalidRequest.Wrap("amount exceeds maximum allowed")
	}
	
	return nil
}

// ValidatePaymentTimeout validates payment timeout parameters
// ValidatePaymentTimeout validates payment timeout parameters
func ValidatePaymentTimeout(timeoutBlocks uint64, currentHeight int64) error {
	if timeoutBlocks == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("payment timeout cannot be zero")
	}
	
	const maxTimeoutBlocks = 86400 // ~6 days at 6s/block
	if timeoutBlocks > maxTimeoutBlocks {
		return sdkerrors.ErrInvalidRequest.Wrapf("payment timeout cannot exceed %d blocks", maxTimeoutBlocks)
	}
	
	return nil
}

// Merchant tier calculation and fee logic

// CalculateMerchantTier determines merchant tier based on stake amount
func CalculateMerchantTier(stakeAmount math.Int) MerchantTier {
	if stakeAmount.GTE(math.NewInt(TierGoldThreshold)) {
		return MerchantTierGold
	}
	if stakeAmount.GTE(math.NewInt(TierSilverThreshold)) {
		return MerchantTierSilver
	}
	return MerchantTierBronze
}

// CalculateTransactionFee calculates fee with tier discount
// basePercent is in decimal form (e.g., 0.001 for 0.1%)
func CalculateTransactionFee(amount math.Int, basePercent math.LegacyDec, tier MerchantTier) math.Int {
	// Apply tier discount
	discount := math.LegacyZeroDec()
	switch tier {
	case MerchantTierGold:
		discount = math.LegacyNewDecWithPrec(50, 2) // 50% discount (0.50)
	case MerchantTierSilver:
		discount = math.LegacyNewDecWithPrec(25, 2) // 25% discount (0.25)
	case MerchantTierBronze:
		discount = math.LegacyZeroDec() // 0% discount
	}
	
	// Calculate effective fee rate
	effectivePercent := basePercent.Mul(math.LegacyOneDec().Sub(discount))
	
	// Calculate fee amount
	feeAmount := math.LegacyNewDecFromInt(amount).Mul(effectivePercent)
	return feeAmount.TruncateInt()
}
