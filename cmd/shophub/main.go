// Komanda shophub pokreće ShopHub HTTP API server.
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Handler("shophub"))
	mux.HandleFunc("GET /readyz", health.Handler("shophub"))
	authHandler.Register(mux)

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