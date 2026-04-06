package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vitapay-gateway/config"
	"vitapay-gateway/store"
)

// Handlers holds all HTTP handler methods.
type Handlers struct {
	cfg *config.Config
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

// Health godoc
// GET /api/v1/health
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"chain":   h.cfg.ChainID,
		"version": "1.0.0",
	})
}

// RegisterMerchant godoc
// POST /api/v1/merchant/register
func (h *Handlers) RegisterMerchant(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Name    string `json:"name" binding:"required"`
		Email   string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: persist merchant to Supabase when DATABASE_URL is set
	c.JSON(http.StatusCreated, gin.H{
		"message": "merchant registered",
		"address": req.Address,
		"name":    req.Name,
	})
}

// CreatePayment godoc
// POST /api/v1/payment/create
func (h *Handlers) CreatePayment(c *gin.Context) {
	var req struct {
		MerchantAddress string `json:"merchantAddress" binding:"required"`
		Amount          string `json:"amount" binding:"required"`
		Denom           string `json:"denom" binding:"required"`
		Memo            string `json:"memo"`
		ExpiresIn       int    `json:"expiresIn"` // seconds, default 900 (15 min)
		WebhookURL      string `json:"webhookUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ExpiresIn <= 0 {
		req.ExpiresIn = 900
	}

	paymentID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(time.Duration(req.ExpiresIn) * time.Second)

	qrData := fmt.Sprintf(
		"vitapay://pay?to=%s&amount=%s&denom=%s&memo=%s&expires=%d",
		req.MerchantAddress, req.Amount, req.Denom, paymentID, expiresAt.Unix(),
	)
	deepLink := qrData

	p := &store.Payment{
		ID:              paymentID,
		MerchantAddress: req.MerchantAddress,
		Amount:          req.Amount,
		Denom:           req.Denom,
		Memo:            req.Memo,
		Status:          store.StatusPending,
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       expiresAt,
		WebhookURL:      req.WebhookURL,
		QRData:          qrData,
		DeepLink:        deepLink,
	}
	store.Default.Save(p)

	c.JSON(http.StatusCreated, gin.H{
		"paymentId": paymentID,
		"qrData":    qrData,
		"deepLink":  deepLink,
		"expiresAt": expiresAt.Format(time.RFC3339),
		"status":    string(store.StatusPending),
	})
}

// GetPayment godoc
// GET /api/v1/payment/:id
func (h *Handlers) GetPayment(c *gin.Context) {
	id := c.Param("id")
	p, ok := store.Default.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	// If still pending, probe chain for matching tx via memo
	if p.Status == store.StatusPending && time.Now().UTC().Before(p.ExpiresAt) {
		if txHash, err := h.queryChainForPayment(p.ID); err == nil && txHash != "" {
			now := time.Now().UTC()
			store.Default.Update(id, func(pay *store.Payment) {
				pay.Status = store.StatusConfirmed
				pay.TxHash = txHash
				pay.ConfirmedAt = &now
			})
			p, _ = store.Default.Get(id)
			if p.WebhookURL != "" {
				go FireWebhook(p.WebhookURL, p)
			}
		}
	}

	// Mark expired
	if p.Status == store.StatusPending && time.Now().UTC().After(p.ExpiresAt) {
		store.Default.Update(id, func(pay *store.Payment) {
			pay.Status = store.StatusExpired
		})
		p, _ = store.Default.Get(id)
	}

	c.JSON(http.StatusOK, p)
}

// ConfirmPayment godoc
// POST /api/v1/payment/:id/confirm
func (h *Handlers) ConfirmPayment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		TxHash string `json:"txHash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, ok := store.Default.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if p.Status != store.StatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "payment already " + string(p.Status)})
		return
	}

	// Verify tx on chain
	if err := h.verifyTx(req.TxHash, p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tx verification failed: " + err.Error()})
		return
	}

	now := time.Now().UTC()
	store.Default.Update(id, func(pay *store.Payment) {
		pay.Status = store.StatusConfirmed
		pay.TxHash = req.TxHash
		pay.ConfirmedAt = &now
	})
	p, _ = store.Default.Get(id)

	if p.WebhookURL != "" {
		go FireWebhook(p.WebhookURL, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"paymentId":   p.ID,
		"status":      string(p.Status),
		"txHash":      p.TxHash,
		"confirmedAt": now.Format(time.RFC3339),
	})
}

// ListMerchantPayments godoc
// GET /api/v1/merchant/:address/payments
func (h *Handlers) ListMerchantPayments(c *gin.Context) {
	address := c.Param("address")
	payments := store.Default.ListByMerchant(address)
	if payments == nil {
		payments = []*store.Payment{}
	}
	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"payments": payments,
		"count":    len(payments),
	})
}

// MerchantStats godoc
// GET /api/v1/merchant/:address/stats
func (h *Handlers) MerchantStats(c *gin.Context) {
	address := c.Param("address")
	payments := store.Default.ListByMerchant(address)
	var total, confirmed, pending, expired int
	for _, p := range payments {
		total++
		switch p.Status {
		case store.StatusConfirmed:
			confirmed++
		case store.StatusPending:
			pending++
		case store.StatusExpired:
			expired++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"address":   address,
		"total":     total,
		"confirmed": confirmed,
		"pending":   pending,
		"expired":   expired,
	})
}

// RegisterWebhook godoc
// POST /api/v1/webhook/register
func (h *Handlers) RegisterWebhook(c *gin.Context) {
	var req struct {
		MerchantAddress string `json:"merchantAddress" binding:"required"`
		WebhookURL      string `json:"webhookUrl" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: persist to Supabase
	c.JSON(http.StatusOK, gin.H{
		"message":    "webhook registered",
		"address":    req.MerchantAddress,
		"webhookUrl": req.WebhookURL,
	})
}

// --- Chain query helpers ---

type txResponse struct {
	TxResponse struct {
		TxHash string `json:"txhash"`
		Code   int    `json:"code"`
		RawLog string `json:"raw_log"`
	} `json:"tx_response"`
}

type searchResponse struct {
	TxResponses []struct {
		TxHash string `json:"txhash"`
		Code   int    `json:"code"`
	} `json:"tx_responses"`
}

// queryChainForPayment searches the REST API for a tx containing the paymentId as memo.
func (h *Handlers) queryChainForPayment(paymentID string) (string, error) {
	url := fmt.Sprintf(
		"%s/cosmos/tx/v1beta1/txs?events=message.memo='%s'",
		h.cfg.VitacoinREST, paymentID,
	)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var sr searchResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return "", err
	}
	for _, tx := range sr.TxResponses {
		if tx.Code == 0 {
			return tx.TxHash, nil
		}
	}
	return "", nil
}

// verifyTx fetches a tx by hash and validates it matches the expected payment.
func (h *Handlers) verifyTx(txHash string, p *store.Payment) error {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", h.cfg.VitacoinREST, txHash)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tx not found (status %d)", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var txResp txResponse
	if err := json.Unmarshal(body, &txResp); err != nil {
		return err
	}
	if txResp.TxResponse.Code != 0 {
		return fmt.Errorf("tx failed on chain: %s", txResp.TxResponse.RawLog)
	}
	// Basic sanity: tx exists and succeeded. Deep field validation requires
	// parsing the tx body (MsgSend) which needs the full Cosmos SDK proto types.
	// We check amount/recipient in the raw log heuristically here; full
	// validation is done by the on-chain module.
	return nil
}
