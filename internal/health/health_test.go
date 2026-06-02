package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Handler("shophub").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("očekivan status 200, dobijen %d", rec.Code)
	}

	var got Response
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("nevalidan JSON: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("očekivan status=ok, dobijen %q", got.Status)
	}
	if got.Service != "shophub" {
		t.Errorf("očekivan service=shophub, dobijen %q", got.Service)
	}
}
