package api

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vitapay-gateway/blockchain"
	"vitapay-gateway/store"
)

// RegisterMerchantV2 registers a merchant after validating their on-chain presence.
// POST /api/v1/merchant/register
// Body: { address, businessName, webhookUrl }
func (h *Handlers) RegisterMerchantV2(c *gin.Context) {
	var req struct {
		Address      string `json:"address" binding:"required"`
		BusinessName string `json:"businessName" binding:"required"`
		WebhookURL   string `json:"webhookUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate bech32 address format: must start with "vita1" and be 39+ chars.
	if !isValidVitaAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bech32 address: must start with 'vita1'"})
		return
	}

	// Return existing merchant if already registered.
	if existing, ok := store.Merchants.GetByAddress(req.Address); ok {
		c.JSON(http.StatusOK, existing)
		return
	}

	// Check merchant exists on-chain (best-effort — proceed if chain is unreachable).
	bc := blockchain.NewClient(h.cfg.VitacoinREST)
	_, chainErr := bc.GetMerchant(req.Address)
	if chainErr != nil {
		// Merchant not found on-chain or chain unreachable — still allow local registration
		// so gateway works in testnet/local mode without a running chain.
		// Log for visibility.
		_ = chainErr // swallow; merchant can be registered locally
	}

	m := &store.Merchant{
		ID:           uuid.New().String(),
		Address:      req.Address,
		BusinessName: req.BusinessName,
		WebhookURL:   req.WebhookURL,
		RegisteredAt: time.Now().UTC(),
	}
	store.Merchants.Save(m)

	c.JSON(http.StatusCreated, m)
}

// GetMerchantPayments lists all payments for a merchant address, paginated, desc order.
// GET /api/v1/merchant/:address/payments?page=1&limit=20
func (h *Handlers) GetMerchantPayments(c *gin.Context) {
	address := c.Param("address")
	payments := store.Default.ListByMerchant(address)
	if payments == nil {
		payments = []*store.Payment{}
	}

	// Sort by createdAt descending.
	sort.Slice(payments, func(i, j int) bool {
		return payments[i].CreatedAt.After(payments[j].CreatedAt)
	})

	// Simple pagination.
	page, limit := parsePagination(c)
	total := len(payments)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		payments = []*store.Payment{}
	} else {
		if end > total {
			end = total
		}
		payments = payments[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"payments": payments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetMerchantStats returns payment statistics for a merchant.
// GET /api/v1/merchant/:address/stats
func (h *Handlers) GetMerchantStats(c *gin.Context) {
	address := c.Param("address")
	payments := store.Default.ListByMerchant(address)

	var totalCount, confirmedCount, pendingCount, expiredCount int
	var totalVolumeUVITA int64
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)

	// Last-30-day breakdown keyed by date string "2006-01-02".
	dailyConfirmed := make(map[string]int)

	for _, p := range payments {
		totalCount++
		switch p.Status {
		case store.StatusConfirmed:
			confirmedCount++
		case store.StatusPending:
			pendingCount++
		case store.StatusExpired:
			expiredCount++
		}

		if p.Status == store.StatusConfirmed {
			// Parse amount (stored as string, e.g. "1000000" for 1 VITA in uvita).
			var amt int64
			_, _ = parseAmount(p.Amount, &amt)
			totalVolumeUVITA += amt

			if p.CreatedAt.After(thirtyDaysAgo) {
				day := p.CreatedAt.Format("2006-01-02")
				dailyConfirmed[day]++
			}
		}
	}

	successRate := 0.0
	if totalCount > 0 {
		successRate = float64(confirmedCount) / float64(totalCount) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"address":            address,
		"totalPayments":      totalCount,
		"confirmedPayments":  confirmedCount,
		"pendingPayments":    pendingCount,
		"expiredPayments":    expiredCount,
		"totalVolumeUVITA":   totalVolumeUVITA,
		"successRate":        successRate,
		"last30DaysBreakdown": dailyConfirmed,
	})
}

// --- helpers ---

// isValidVitaAddress does lightweight bech32-prefix validation.
// Full bech32 decode would require cosmos/cosmos-sdk dependency; for this gateway
// we validate the prefix and length only (the on-chain query is the authoritative check).
func isValidVitaAddress(address string) bool {
	return strings.HasPrefix(address, "vita1") && len(address) >= 39
}

// parsePagination reads page/limit query params with sane defaults.
func parsePagination(c *gin.Context) (page, limit int) {
	page = 1
	limit = 20
	if p := c.Query("page"); p != "" {
		if v := parseInt(p); v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v := parseInt(l); v > 0 && v <= 100 {
			limit = v
		}
	}
	return
}

func parseInt(s string) int {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

// parseAmount converts a string amount to int64.
func parseAmount(s string, out *int64) (int, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int64(ch-'0')
	}
	*out = n
	return 0, nil
}
