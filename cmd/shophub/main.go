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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shophub-platform/shophub/internal/auth"
	"github.com/shophub-platform/shophub/internal/config"
	"github.com/shophub-platform/shophub/internal/database"
	"github.com/shophub-platform/shophub/internal/health"
	"github.com/shophub-platform/shophub/internal/httpapi"
	"github.com/shophub-platform/shophub/internal/k8s"
	"github.com/shophub-platform/shophub/internal/shops"
	"github.com/shophub-platform/shophub/pkg/version"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("konekcija ka bazi nije uspela: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migracija nije uspela: %v", err)
	}

	tm := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authHandler := httpapi.NewAuthHandler(db, tm)

	// Kubernetes orkestrator (client-go). Ako klaster nije dostupan, pada na
	// no-op kako bi REST API i dalje radio lokalno (CR-ovi se ne kreiraju).
	var orch shops.Orchestrator
	if client, err := k8s.NewClient(cfg.KubeconfigPath); err != nil {
		log.Printf("Kubernetes nije dostupan (%v) — Shop CR orkestracija je onemogućena", err)
		orch = shops.NewNoopOrchestrator()
	} else {
		log.Printf("Kubernetes orkestrator aktivan (namespace=%s)", cfg.ShopNamespace)
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
	authHandler.Register(mux)
	shopHandler.Register(mux)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("ShopHub %s sluša na :%s", version.Version, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server greška: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown greška: %v", err)
	}
	log.Println("server zaustavljen")
}