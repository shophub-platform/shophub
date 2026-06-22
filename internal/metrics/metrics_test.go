package metrics_test

import (
	"testing"

	"github.com/shophub-platform/shophub/internal/metrics"
)

func TestRegisterNopanic(t *testing.T) {
	metrics.Register()
}
