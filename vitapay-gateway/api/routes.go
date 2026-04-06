package api

import (
	"github.com/gin-gonic/gin"
	"vitapay-gateway/config"
	"vitapay-gateway/middleware"
)

// RegisterRoutes wires all HTTP routes onto the gin engine.
func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	h := NewHandlers(cfg)

	// Public
	r.GET("/api/v1/health", h.Health)

	v1 := r.Group("/api/v1")

	// Merchant routes (JWT-protected)
	merchant := v1.Group("/merchant")
	merchant.Use(middleware.JWT(cfg.JWTSecret))
	{
		merchant.POST("/register", h.RegisterMerchant)
		merchant.GET("/:address/payments", h.ListMerchantPayments)
		merchant.GET("/:address/stats", h.MerchantStats)
	}

	// Payment routes — create is public (merchants call with their own auth at app layer)
	payment := v1.Group("/payment")
	{
		payment.POST("/create", h.CreatePayment)
		payment.GET("/:id", h.GetPayment)
		payment.POST("/:id/confirm", h.ConfirmPayment)
	}

	// Webhook management (JWT-protected)
	webhook := v1.Group("/webhook")
	webhook.Use(middleware.JWT(cfg.JWTSecret))
	{
		webhook.POST("/register", h.RegisterWebhook)
	}
}
