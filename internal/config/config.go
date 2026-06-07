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

	// Kubernetes / Shop CR (FZ 2.2)
	KubeconfigPath   string // putanja do kubeconfig-a (prazno = in-cluster ili ~/.kube/config)
	ShopNamespace    string // namespace u koji idu Shop CR-ovi
	DefaultShopImage string // podrazumevana container slika prodavnice
	ShopURLTemplate  string // %s = ime CR-a, npr. "http://%s.shophub.local"
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

		KubeconfigPath:   getenv("KUBECONFIG", ""),
		ShopNamespace:    getenv("SHOP_NAMESPACE", "default"),
		DefaultShopImage: getenv("SHOP_DEFAULT_IMAGE", "ghcr.io/shophub-platform/shop:latest"),
		ShopURLTemplate:  getenv("SHOP_URL_TEMPLATE", "http://%s.shophub.local"),
	}
}
