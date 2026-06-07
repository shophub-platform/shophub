//go:build integration

// Integracioni testovi auth + shops endpoint-a sa pravom PostgreSQL bazom
// (Testcontainers) i pravim HTTP zahtevima (FZ 2.3).
//
// Pokretanje:  go test -tags=integration ./internal/httpapi/ -run Integration -v
// Zahteva Docker.
package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	"github.com/shophub-platform/shophub/internal/auth"
	"github.com/shophub-platform/shophub/internal/database"
	"github.com/shophub-platform/shophub/internal/httpapi"
	"github.com/shophub-platform/shophub/internal/shops"
)

func startPostgres(t *testing.T) (dsn string, cleanup func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "shophub",
			"POSTGRES_PASSWORD": "shophub",
			"POSTGRES_DB":       "shophub",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(90 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	require.NoError(t, err)

	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn = fmt.Sprintf("host=%s user=shophub password=shophub dbname=shophub port=%s sslmode=disable",
		host, port.Port())
	return dsn, func() { _ = c.Terminate(ctx) }
}

func newTestServer(t *testing.T, dsn string) *httptest.Server {
	t.Helper()
	db := mustConnect(t, dsn)
	require.NoError(t, database.Migrate(db))

	tm := auth.NewTokenManager("integration-secret", 15*time.Minute, 24*time.Hour)
	authHandler := httpapi.NewAuthHandler(db, tm)
	svc := shops.NewService(shops.NewGormRepository(db), shops.NewNoopOrchestrator(), shops.Options{
		Namespace: "default", DefaultImage: "img:1", URLTemplate: "http://%s.shophub.test",
	})
	shopHandler := httpapi.NewShopHandler(svc, tm)

	mux := http.NewServeMux()
	authHandler.Register(mux)
	shopHandler.Register(mux)
	return httptest.NewServer(mux)
}

func mustConnect(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	var lastErr error
	for i := 0; i < 10; i++ {
		db, err := database.Connect(dsn)
		if err == nil {
			return db
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	t.Fatalf("konekcija ka bazi nije uspela: %v", lastErr)
	return nil
}

func postJSON(t *testing.T, url, token string, body any) (*http.Response, []byte) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, data
}

func TestIntegrationAuthAndShops(t *testing.T) {
	dsn, cleanup := startPostgres(t)
	defer cleanup()
	srv := newTestServer(t, dsn)
	defer srv.Close()

	// 1) registracija
	resp, body := postJSON(t, srv.URL+"/api/v1/auth/register", "", map[string]string{
		"email": "vesna@b.com", "password": "supersecret", "displayName": "Vesna",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))

	// 2) login
	resp, body = postJSON(t, srv.URL+"/api/v1/auth/login", "", map[string]string{
		"email": "vesna@b.com", "password": "supersecret",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var tok struct {
		AccessToken string `json:"accessToken"`
	}
	require.NoError(t, json.Unmarshal(body, &tok))
	require.NotEmpty(t, tok.AccessToken)

	// 3) kreiranje prodavnice
	resp, body = postJSON(t, srv.URL+"/api/v1/shops", tok.AccessToken, map[string]string{
		"name": "moja-radnja", "availability": "high", "walletAddress": "0xabc", "databaseType": "postgres",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var created struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	require.NoError(t, json.Unmarshal(body, &created))
	require.NotEmpty(t, created.ID)

	// 4) lista
	listReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/shops", nil)
	listReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	listResp, err := http.DefaultClient.Do(listReq)
	require.NoError(t, err)
	listBody, _ := io.ReadAll(listResp.Body)
	_ = listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(listBody, &list))
	require.Len(t, list, 1)

	// 5) brisanje
	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/shops/"+created.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err)
	_ = delResp.Body.Close()
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// 6) nakon brisanja -> 404
	getReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/shops/"+created.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	_ = getResp.Body.Close()
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
