package database_test

import (
	"testing"

	"github.com/shophub-platform/shophub/internal/database"
)

func TestConnectInvalidDSN(t *testing.T) {
	_, err := database.Connect("invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}
