// Package health pruža health/readiness HTTP handler-e.
package health

import (
	"encoding/json"
	"net/http"
)

// Response je telo health odgovora.
type Response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Handler vraća 200 OK sa JSON statusom. Koristi se za /healthz i /readyz.
func Handler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{Status: "ok", Service: service})
	}
}
