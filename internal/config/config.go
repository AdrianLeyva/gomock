// Package config loads runtime configuration from environment variables.
package config

import "os"

// Config holds runtime configuration for the server.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
	// DataDir is the directory scanned for per-type JSON data files.
	DataDir string
}

// Load reads configuration from environment variables, applying defaults
// when they are unset: PORT defaults to "8080", DATA_DIR to "./data".
func Load() Config {
	return Config{
		Port:    getEnv("PORT", "8080"),
		DataDir: getEnv("DATA_DIR", "./data"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
