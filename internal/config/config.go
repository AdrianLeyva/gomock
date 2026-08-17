// Package config loads runtime configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration for the server.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
	// DataDir is the directory scanned for per-type JSON data files.
	DataDir string
	// RateLimitRPS is the sustained requests/second allowed per client IP.
	// Zero or negative disables rate limiting.
	RateLimitRPS int
	// RateLimitBurst is the maximum burst of requests allowed per client IP.
	RateLimitBurst int
}

// Load reads configuration from environment variables, applying defaults when
// they are unset: PORT "8080", DATA_DIR "./data", RATE_LIMIT_RPS 10,
// RATE_LIMIT_BURST 20.
func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DataDir:        getEnv("DATA_DIR", "./data"),
		RateLimitRPS:   getEnvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
	}
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
