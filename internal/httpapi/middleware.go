package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/shophub-platform/shophub/internal/auth"
)

type ctxKey string

const (
	ctxUserID ctxKey = "userID"
	ctxEmail  ctxKey = "email"
)

// authTokenManager je minimalni interfejs koji RequireAuth zahteva.
// *auth.TokenManager ga zadovoljava, a u testovima se lako mock-uje.
type authTokenManager interface {
	Parse(tokenString string, expected auth.TokenType) (*auth.Claims, error)
}

func RequireAuth(tm authTokenManager, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "nedostaje bearer token")
			return
		}
		claims, err := tm.Parse(parts[1], auth.AccessToken)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "nevalidan token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID.String())
		ctx = context.WithValue(ctx, ctxEmail, claims.Email)
		next(w, r.WithContext(ctx))
	}
}

func userIDFromContext(r *http.Request) string {
	v, _ := r.Context().Value(ctxUserID).(string)
	return v
}

func emailFromContext(r *http.Request) string {
	v, _ := r.Context().Value(ctxEmail).(string)
	return v
}
