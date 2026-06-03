package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/shophub-platform/shophub/internal/auth"
	"github.com/shophub-platform/shophub/internal/models"
)

// AuthHandler grupiše auth endpoint-e (FZ 1.1: registracija, login, refresh).
type AuthHandler struct {
	DB     *gorm.DB
	Tokens *auth.TokenManager
}

// NewAuthHandler kreira AuthHandler.
func NewAuthHandler(db *gorm.DB, tm *auth.TokenManager) *AuthHandler {
	return &AuthHandler{DB: db, Tokens: tm}
}

// Register registruje endpoint-e na dati mux.
func (h *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /api/v1/auth/logout", RequireAuth(h.Tokens, h.handleLogout))
	mux.HandleFunc("GET /api/v1/me", RequireAuth(h.Tokens, h.handleMe))
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
}

func (h *AuthHandler) issueTokens(u models.User) (*tokenResponse, error) {
	access, err := h.Tokens.NewAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, err
	}
	refresh, err := h.Tokens.NewRefreshToken(u.ID, u.Email, u.RefreshTokenVersion)
	if err != nil {
		return nil, err
	}
	return &tokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}, nil
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "nevalidan JSON")
		return
	}
	if req.Email == "" || len(req.Password) < 8 || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "email, displayName i lozinka (min 8 karaktera) su obavezni")
		return
	}

	var existing models.User
	err := h.DB.Where("email = ?", req.Email).First(&existing).Error
	if err == nil {
		writeError(w, http.StatusConflict, "email je već registrovan")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(w, http.StatusInternalServerError, "greška baze")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da heširam lozinku")
		return
	}

	user := models.User{Email: req.Email, PasswordHash: hash, DisplayName: req.DisplayName}
	if err := h.DB.Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da kreiram korisnika")
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da izdam tokene")
		return
	}
	writeJSON(w, http.StatusCreated, tokens)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "nevalidan JSON")
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		writeError(w, http.StatusUnauthorized, "nevalidni kredencijali")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "nevalidni kredencijali")
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da izdam tokene")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "nevalidan JSON")
		return
	}

	claims, err := h.Tokens.Parse(req.RefreshToken, auth.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "nevalidan refresh token")
		return
	}

	var user models.User
	if err := h.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
		writeError(w, http.StatusUnauthorized, "korisnik nije pronađen")
		return
	}
	// Odbij refresh tokene izdate pre tekuće verzije (npr. posle logout-a).
	if claims.Version != user.RefreshTokenVersion {
		writeError(w, http.StatusUnauthorized, "refresh token je poništen")
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da izdam tokene")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("refresh_token_version", gorm.Expr("refresh_token_version + 1")).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "ne mogu da izvršim logout")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "odjavljen"})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"userID": userIDFromContext(r),
		"email":  emailFromContext(r),
	})
}
