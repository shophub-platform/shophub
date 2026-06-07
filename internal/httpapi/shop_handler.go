package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/shophub-platform/shophub/internal/shops"
)

// ShopHandler izlaže REST endpoint-e za prodavnice (FZ 2.1).
type ShopHandler struct {
	svc    shops.Service
	tokens authTokenManager
}

// NewShopHandler kreira ShopHandler.
func NewShopHandler(svc shops.Service, tm authTokenManager) *ShopHandler {
	return &ShopHandler{svc: svc, tokens: tm}
}

// Register registruje shop endpoint-e (svi zahtevaju Bearer token).
func (h *ShopHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/shops", RequireAuth(h.tokens, h.handleList))
	mux.HandleFunc("POST /api/v1/shops", RequireAuth(h.tokens, h.handleCreate))
	mux.HandleFunc("GET /api/v1/shops/{id}", RequireAuth(h.tokens, h.handleGet))
	mux.HandleFunc("PATCH /api/v1/shops/{id}", RequireAuth(h.tokens, h.handleUpdate))
	mux.HandleFunc("DELETE /api/v1/shops/{id}", RequireAuth(h.tokens, h.handleDelete))
	mux.HandleFunc("GET /api/v1/shops/{id}/url", RequireAuth(h.tokens, h.handleURL))
}

// --- DTO-ovi ---

type createShopRequest struct {
	Name          string `json:"name"`
	Availability  string `json:"availability"`
	WalletAddress string `json:"walletAddress"`
	DatabaseType  string `json:"databaseType"`
	Image         string `json:"image,omitempty"`
}

type updateShopRequest struct {
	Availability  *string `json:"availability,omitempty"`
	WalletAddress *string `json:"walletAddress,omitempty"`
	DatabaseType  *string `json:"databaseType,omitempty"`
}

type shopResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Availability  string    `json:"availability"`
	WalletAddress string    `json:"walletAddress"`
	DatabaseType  string    `json:"databaseType"`
	Image         string    `json:"image"`
	Status        string    `json:"status"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func toResponse(v shops.View) shopResponse {
	return shopResponse{
		ID:            v.Shop.ID.String(),
		Name:          v.Shop.Name,
		Availability:  string(v.Shop.Availability),
		WalletAddress: v.Shop.WalletAddr,
		DatabaseType:  string(v.Shop.DatabaseType),
		Image:         v.Shop.Image,
		Status:        v.Status,
		URL:           v.URL,
		CreatedAt:     v.Shop.CreatedAt,
		UpdatedAt:     v.Shop.UpdatedAt,
	}
}

// --- helpers ---

func ownerID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(userIDFromContext(r))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func pathID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// mapServiceError prevodi servisne greške na HTTP statuse.
func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shops.ErrNotFound):
		writeError(w, http.StatusNotFound, "prodavnica nije pronađena")
	case errors.Is(err, shops.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "interna greška")
	}
}

// --- handleri ---

// handleList godoc
// @Summary  Lista korisnikovih prodavnica
// @Tags     shops
// @Produce  json
// @Security BearerAuth
// @Success  200 {array} shopResponse
// @Router   /api/v1/shops [get]
func (h *ShopHandler) handleList(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	views, err := h.svc.List(r.Context(), uid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	out := make([]shopResponse, 0, len(views))
	for _, v := range views {
		out = append(out, toResponse(v))
	}
	writeJSON(w, http.StatusOK, out)
}

// handleCreate godoc
// @Summary  Kreira prodavnicu (Shop CR + meta-podaci u DB)
// @Tags     shops
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body createShopRequest true "Podaci prodavnice"
// @Success  201 {object} shopResponse
// @Failure  400 {object} map[string]string
// @Router   /api/v1/shops [post]
func (h *ShopHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	var req createShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "nevalidan JSON")
		return
	}
	v, err := h.svc.Create(r.Context(), uid, shops.CreateInput{
		Name:          req.Name,
		Availability:  req.Availability,
		WalletAddress: req.WalletAddress,
		DatabaseType:  req.DatabaseType,
		Image:         req.Image,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(*v))
}

// handleGet godoc
// @Summary  Detalji prodavnice (status iz CR phase)
// @Tags     shops
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Shop ID (UUID)"
// @Success  200 {object} shopResponse
// @Failure  404 {object} map[string]string
// @Router   /api/v1/shops/{id} [get]
func (h *ShopHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "nevalidan id")
		return
	}
	v, err := h.svc.Get(r.Context(), uid, id)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*v))
}

// handleUpdate godoc
// @Summary  Menja konfiguraciju (availability, wallet, database)
// @Tags     shops
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path string            true "Shop ID (UUID)"
// @Param    body body updateShopRequest true "Polja za izmenu"
// @Success  200 {object} shopResponse
// @Failure  404 {object} map[string]string
// @Router   /api/v1/shops/{id} [patch]
func (h *ShopHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "nevalidan id")
		return
	}
	var req updateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "nevalidan JSON")
		return
	}
	v, err := h.svc.Update(r.Context(), uid, id, shops.UpdateInput{
		Availability:  req.Availability,
		WalletAddress: req.WalletAddress,
		DatabaseType:  req.DatabaseType,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*v))
}

// handleDelete godoc
// @Summary  Briše prodavnicu (operator čisti resurse)
// @Tags     shops
// @Security BearerAuth
// @Param    id path string true "Shop ID (UUID)"
// @Success  204 "No Content"
// @Failure  404 {object} map[string]string
// @Router   /api/v1/shops/{id} [delete]
func (h *ShopHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "nevalidan id")
		return
	}
	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleURL godoc
// @Summary  Javni URL prodavnice
// @Tags     shops
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Shop ID (UUID)"
// @Success  200 {object} map[string]string
// @Failure  404 {object} map[string]string
// @Router   /api/v1/shops/{id}/url [get]
func (h *ShopHandler) handleURL(w http.ResponseWriter, r *http.Request) {
	uid, ok := ownerID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "nevalidan identitet")
		return
	}
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "nevalidan id")
		return
	}
	url, err := h.svc.URL(r.Context(), uid, id)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
