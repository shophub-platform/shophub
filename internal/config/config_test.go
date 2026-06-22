package config_test

import (
	"testing"

	"github.com/shophub-platform/shophub/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg := config.Load()
	if cfg.Port == "" {
		t.Error("expected non-empty default port")
	}
}
