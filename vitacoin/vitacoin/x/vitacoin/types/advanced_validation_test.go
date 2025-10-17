package types_test

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

func TestAdvancedValidation_BusinessName(t *testing.T) {
	tests := []struct {
		name        string
		businessName string
		expectError bool
		errorMsg    string
	}{
		{
			name:         "valid business name",
			businessName: "Valid Business Co.",
			expectError:  false,
		},
		{
			name:         "valid business name with numbers",
			businessName: "Tech Company 123",
			expectError:  false,
		},
		{
			name:         "valid business name with special chars",
			businessName: "ABC Corp & Associates (LLC)",
			expectError:  false,
		},
		{
			name:         "empty business name",
			businessName: "",
			expectError:  true,
			errorMsg:     "cannot be empty",
		},
		{
			name:         "business name too short",
			businessName: "AB",
			expectError:  true,
			errorMsg:     "must be at least",
		},
		{
			name:         "business name too long",
			businessName: strings.Repeat("A", 101),
			expectError:  true,
			errorMsg:     "cannot exceed",
		},
		{
			name:         "business name with control characters",
			businessName: "Invalid\x00Business",
			expectError:  true,
			errorMsg:     "control characters",
		},
		{
			name:         "business name with leading whitespace",
			businessName: " Invalid Business",
			expectError:  true,
			errorMsg:     "whitespace",
		},
		{
			name:         "business name with trailing whitespace",
			businessName: "Invalid Business ",
			expectError:  true,
			errorMsg:     "whitespace",
		},
		{
			name:         "business name with consecutive spaces",
			businessName: "Invalid  Business",
			expectError:  true,
			errorMsg:     "consecutive spaces",
		},
		{
			name:         "business name with invalid characters",
			businessName: "Business@#$%",
			expectError:  true,
			errorMsg:     "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateBusinessName(tt.businessName)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_PaymentAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      math.Int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid amount",
			amount:      math.NewInt(1e18), // 1 VITA
			expectError: false,
		},
		{
			name:        "minimum valid amount",
			amount:      math.NewInt(1e15), // 0.001 VITA (updated minimum)
			expectError: false,
		},
		{
			name:        "zero amount",
			amount:      math.ZeroInt(),
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name:        "negative amount",
			amount:      math.NewInt(-1000),
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name:        "amount too small",
			amount:      math.NewInt(1e14), // Less than minimum (1e15)
			expectError: true,
			errorMsg:    "must be at least",
		},
		{
			name:        "stake amount too large",
			amount:      math.NewInt(1000000).Mul(math.NewInt(1000000000000000000)).Add(math.NewInt(1)), // > 1M VITA
			expectError: true,
			errorMsg:    "maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidatePaymentAmount(tt.amount)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_VaultAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      math.Int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid vault amount",
			amount:      math.NewInt(1e18), // 1 VITA
			expectError: false,
		},
		{
			name:        "large valid amount",
			amount:      math.NewInt(1000000).Mul(math.NewInt(1000000000000000000)), // 1M VITA
			expectError: false,
		},
		{
			name:        "zero amount",
			amount:      math.ZeroInt(),
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name:        "amount below minimum",
			amount:      math.NewInt(1e17), // 0.1 VITA
			expectError: true,
			errorMsg:    "must be at least",
		},
		{
			name:        "amount too large",
			amount:      math.NewInt(10000000).Mul(math.NewInt(1000000000000000000)).Add(math.NewInt(1)), // > 10M VITA
			expectError: true,
			errorMsg:    "maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateVaultAmount(tt.amount)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_LockDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    uint64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid duration",
			duration:    1000,
			expectError: false,
		},
		{
			name:        "minimum duration",
			duration:    1,
			expectError: false,
		},
		{
			name:        "maximum duration",
			duration:    5_256_000, // 1 year
			expectError: false,
		},
		{
			name:        "zero duration",
			duration:    0,
			expectError: true,
			errorMsg:    "must be at least",
		},
		{
			name:        "duration too long",
			duration:    5_256_001,
			expectError: true,
			errorMsg:    "cannot exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateLockDuration(tt.duration)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_UnlockHeight(t *testing.T) {
	tests := []struct {
		name          string
		unlockHeight  int64
		currentHeight int64
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid unlock height",
			unlockHeight:  1000,
			currentHeight: 100,
			expectError:   false,
		},
		{
			name:          "unlock height in past",
			unlockHeight:  100,
			currentHeight: 200,
			expectError:   true,
			errorMsg:      "must be in the future",
		},
		{
			name:          "unlock height equal to current",
			unlockHeight:  100,
			currentHeight: 100,
			expectError:   true,
			errorMsg:      "must be in the future",
		},
		{
			name:          "unlock height too high",
			unlockHeight:  100_000_001,
			currentHeight: 100,
			expectError:   true,
			errorMsg:      "cannot exceed",
		},
		{
			name:          "duration too long",
			unlockHeight:  100 + 5_256_001,
			currentHeight: 100,
			expectError:   true,
			errorMsg:      "cannot exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateUnlockHeight(tt.unlockHeight, tt.currentHeight)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_Memo(t *testing.T) {
	tests := []struct {
		name        string
		memo        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid memo",
			memo:        "Payment for services",
			expectError: false,
		},
		{
			name:        "empty memo",
			memo:        "",
			expectError: false,
		},
		{
			name:        "memo with tab and newline",
			memo:        "Line 1\tTab\nLine 2",
			expectError: false,
		},
		{
			name:        "memo too long",
			memo:        strings.Repeat("A", 257),
			expectError: true,
			errorMsg:    "cannot exceed",
		},
		{
			name:        "memo with control characters",
			memo:        "Invalid\x00memo",
			expectError: true,
			errorMsg:    "control characters",
		},
		{
			name:        "memo with invalid UTF-8",
			memo:        string([]byte{0xff, 0xfe}),
			expectError: true,
			errorMsg:    "invalid UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateMemo(tt.memo)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_RewardDistribution(t *testing.T) {
	tests := []struct {
		name        string
		recipients  []string
		amounts     []math.Int
		expectError bool
		errorMsg    string
	}{
		{
			name:       "valid distribution",
			recipients: []string{validAddress1, validAddress2},
			amounts:    []math.Int{math.NewInt(1e18), math.NewInt(2e18)},
			expectError: false,
		},
		{
			name:        "empty recipients",
			recipients:  []string{},
			amounts:     []math.Int{},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "mismatched lengths",
			recipients:  []string{validAddress1, validAddress2},
			amounts:     []math.Int{math.NewInt(1e18)},
			expectError: true,
			errorMsg:    "same length",
		},
		{
			name:        "invalid recipient address",
			recipients:  []string{invalidAddress},
			amounts:     []math.Int{math.NewInt(1e18)},
			expectError: true,
			errorMsg:    "invalid recipient address",
		},
		{
			name:        "duplicate recipients",
			recipients:  []string{validAddress1, validAddress1},
			amounts:     []math.Int{math.NewInt(1e18), math.NewInt(2e18)},
			expectError: true,
			errorMsg:    "duplicate recipient",
		},
		{
			name:        "zero amount",
			recipients:  []string{validAddress1},
			amounts:     []math.Int{math.ZeroInt()},
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name:        "amount too small",
			recipients:  []string{validAddress1},
			amounts:     []math.Int{math.NewInt(1e14)}, // Below minimum
			expectError: true,
			errorMsg:    "too small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateRewardDistribution(tt.recipients, tt.amounts)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_ID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		idType      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid ID",
			id:          "payment-123-abc",
			idType:      "payment",
			expectError: false,
		},
		{
			name:        "empty ID",
			id:          "",
			idType:      "payment",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "ID too long",
			id:          strings.Repeat("A", 1001),
			idType:      "payment",
			expectError: true,
			errorMsg:    "cannot exceed",
		},
		{
			name:        "ID with control characters",
			id:          "payment\x00123",
			idType:      "payment",
			expectError: true,
			errorMsg:    "control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateID(tt.id, tt.idType)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_NoReentrancy(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "safe ID",
			id:          "payment-123-abc",
			expectError: false,
		},
		{
			name:        "ID with path traversal",
			id:          "payment-../etc/passwd",
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "ID with script tag",
			id:          "payment-<script>alert('xss')</script>",
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
		{
			name:        "ID with javascript",
			id:          "payment-javascript:alert(1)",
			expectError: true,
			errorMsg:    "dangerous pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateNoReentrancy(tt.id)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_AddressPair(t *testing.T) {
	tests := []struct {
		name        string
		addr1       string
		addr2       string
		name1       string
		name2       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid address pair",
			addr1:       validAddress1,
			addr2:       validAddress2,
			name1:       "sender",
			name2:       "recipient",
			expectError: false,
		},
		{
			name:        "invalid first address",
			addr1:       invalidAddress,
			addr2:       validAddress2,
			name1:       "sender",
			name2:       "recipient",
			expectError: true,
			errorMsg:    "invalid sender address",
		},
		{
			name:        "invalid second address",
			addr1:       validAddress1,
			addr2:       invalidAddress,
			name1:       "sender",
			name2:       "recipient",
			expectError: true,
			errorMsg:    "invalid recipient address",
		},
		{
			name:        "same addresses",
			addr1:       validAddress1,
			addr2:       validAddress1,
			name1:       "sender",
			name2:       "recipient",
			expectError: true,
			errorMsg:    "cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateAddressPair(tt.addr1, tt.addr2, tt.name1, tt.name2)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdvancedValidation_Reason(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid reason",
			reason:      "Customer requested refund",
			expectError: false,
		},
		{
			name:        "empty reason",
			reason:      "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "reason too long",
			reason:      strings.Repeat("A", 257),
			expectError: true,
			errorMsg:    "cannot exceed",
		},
		{
			name:        "reason with control characters",
			reason:      "Invalid\x00reason",
			expectError: true,
			errorMsg:    "control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateReason(tt.reason)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Benchmark tests for validation performance
func BenchmarkBusinessNameValidation(b *testing.B) {
	businessName := "Test Business Corporation Ltd."
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = types.ValidateBusinessName(businessName)
	}
}

func BenchmarkPaymentAmountValidation(b *testing.B) {
	amount := math.NewInt(1e18)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = types.ValidatePaymentAmount(amount)
	}
}

func BenchmarkRewardDistributionValidation(b *testing.B) {
	recipients := []string{validAddress1, validAddress2}
	amounts := []math.Int{math.NewInt(1e18), math.NewInt(2e18)}
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = types.ValidateRewardDistribution(recipients, amounts)
	}
}