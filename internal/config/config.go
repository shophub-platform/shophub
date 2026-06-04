// Package config učitava konfiguraciju iz okruženja (environment varijabli).
package config

import (
	"os"
	"time"
)

// Config drži runtime konfiguraciju ShopHub servisa.
type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load čita konfiguraciju iz okruženja uz razumne podrazumevane vrednosti.
func Load() *Config {
	return &Config{
		Port: getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL",
			"host=localhost user=shophub password=shophub dbname=shophub port=5432 sslmode=disable"),
		JWTSecret:       getenv("JWT_SECRET", "dev-insecure-secret-promeni-me"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
}
