package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"vitapay-gateway/api"
	"vitapay-gateway/config"
	"vitapay-gateway/middleware"
)

func main() {
	cfg := config.Load()
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())
	api.RegisterRoutes(r, cfg)
	log.Printf("VITAPAY Gateway starting on :%s (chain=%s)", cfg.Port, cfg.ChainID)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
