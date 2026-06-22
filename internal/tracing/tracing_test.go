package tracing_test

import (
	"context"
	"testing"

	"github.com/shophub-platform/shophub/internal/tracing"
	"go.uber.org/zap"
)

func TestSetupNoEndpoint(t *testing.T) {
	// Without OTEL_EXPORTER_OTLP_ENDPOINT set, Setup should return a no-op shutdown func.
	logger := zap.NewNop()
	shutdown := tracing.Setup(context.Background(), logger)
	shutdown()
}
