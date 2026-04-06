package config

import (
	"os"
)

// Config holds all runtime configuration for the VITAPAY Gateway.
// All secrets are read from environment variables — never hardcoded.
type Config struct {
	Port          string
	VitacoinRPC   string // e.g. https://rpc.vitacoin.network
	VitacoinREST  string // e.g. https://api.vitacoin.network
	ChainID       string // vitacoin-1
	ModuleAddress string // gateway module account on-chain
	JWTSecret     string // VITAPAY_JWT_SECRET
	DatabaseURL   string // DATABASE_URL (Supabase postgres connection string)
}

// Load reads configuration from environment variables with sane defaults.
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		VitacoinRPC:   getEnv("VITACOIN_RPC", "https://rpc.vitacoin.network"),
		VitacoinREST:  getEnv("VITACOIN_REST", "https://api.vitacoin.network"),
		ChainID:       getEnv("CHAIN_ID", "vitacoin-1"),
		ModuleAddress: getEnv("MODULE_ADDRESS", ""),
		JWTSecret:     getEnv("VITAPAY_JWT_SECRET", ""),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
