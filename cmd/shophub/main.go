// Komanda shophub pokreće ShopHub HTTP API server.
//
// @title           ShopHub API
// @version         0.1.0
// @description     REST API za upravljanje nalozima i sajtovima prodavnica
// @description     koje shop-operator deploy-uje u Kubernetes preko Shop CRD-a.
// @BasePath        /
// @securityDefinitions.apikey  BearerAuth
// @in              header
// @name            Authorization
// @description     Unesi: "Bearer <access token>"
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/shophub-platform/shophub/internal/auth"
	"github.com/shophub-platform/shophub/internal/config"
	"github.com/shophub-platform/shophub/internal/database"
	"github.com/shophub-platform/shophub/internal/health"
	"github.com/shophub-platform/shophub/internal/httpapi"
	"github.com/shophub-platform/shophub/internal/k8s"
	"github.com/shophub-platform/shophub/internal/metrics"
	"github.com/shophub-platform/shophub/internal/shops"
	"github.com/shophub-platform/shophub/internal/tracing"
	"github.com/shophub-platform/shophub/pkg/version"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg := config.Load()

	metrics.Register()

	shutdownTracing := tracing.Setup(context.Background(), logger)
	defer shutdownTracing()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("konekcija ka bazi nije uspela", zap.Error(err))
	}
	if err := database.Migrate(db); err != nil {
		logger.Fatal("migracija nije uspela", zap.Error(err))
	}

	tm := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authHandler := httpapi.NewAuthHandler(db, tm)

	// Kubernetes orkestrator (client-go). Ako klaster nije dostupan, pada na
	// no-op kako bi REST API i dalje radio lokalno (CR-ovi se ne kreiraju).
	var orch shops.Orchestrator
	if client, err := k8s.NewClient(cfg.KubeconfigPath); err != nil {
		logger.Warn("Kubernetes nije dostupan — Shop CR orkestracija je onemogućena", zap.Error(err))
		orch = shops.NewNoopOrchestrator()
	} else {
		logger.Info("Kubernetes orkestrator aktivan", zap.String("namespace", cfg.ShopNamespace))
		orch = client
	}

	shopSvc := shops.NewService(shops.NewGormRepository(db), orch, shops.Options{
		Namespace:    cfg.ShopNamespace,
		DefaultImage: cfg.DefaultShopImage,
		URLTemplate:  cfg.ShopURLTemplate,
	})
	shopHandler := httpapi.NewShopHandler(shopSvc, tm)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Handler("shophub"))
	mux.HandleFunc("GET /readyz", health.Handler("shophub"))
	// Prometheus scrape endpoint — not counted in metrics.
	mux.Handle("GET /metrics", promhttp.Handler())
	authHandler.Register(mux)
	shopHandler.Register(mux)

	// Wrap mux: OTel tracing outermost, then Prometheus metrics.
	var handler http.Handler = mux
	handler = httpapi.MetricsMiddleware(handler)
	handler = otelhttp.NewHandler(handler, "shophub")

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("ShopHub started", zap.String("version", version.Version), zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server greška", zap.Error(err))
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown greška", zap.Error(err))
	}
	logger.Info("server zaustavljen")
}