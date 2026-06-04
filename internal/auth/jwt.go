package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

var (
	// ErrInvalidToken se vraća za nevalidan ili istekao token.
	ErrInvalidToken = errors.New("nevalidan token")
	// ErrWrongType se vraća kada tip tokena nije očekivan.
	ErrWrongType = errors.New("neočekivan tip tokena")
)

// Claims je JWT payload koji koristi ShopHub.
type Claims struct {
	UserID  uuid.UUID `json:"uid"`
	Email   string    `json:"email"`
	Type    TokenType `json:"typ"`
	Version int       `json:"ver"` // poklapa se sa User.RefreshTokenVersion za refresh tokene
	jwt.RegisteredClaims
}

// TokenManager izdaje i validira JWT tokene.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager kreira novi TokenManager.
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *TokenManager) generate(userID uuid.UUID, email string, t TokenType, version int, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Email:   email,
		Type:    t,
		Version: version,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// NewAccessToken izdaje kratkotrajni access token.
func (m *TokenManager) NewAccessToken(userID uuid.UUID, email string) (string, error) {
	return m.generate(userID, email, AccessToken, 0, m.accessTTL)
}

// NewRefreshToken izdaje dugotrajni refresh token vezan za verziju.
func (m *TokenManager) NewRefreshToken(userID uuid.UUID, email string, version int) (string, error) {
	return m.generate(userID, email, RefreshToken, version, m.refreshTTL)
}

// Parse validira token i potvrđuje da je očekivanog tipa.
func (m *TokenManager) Parse(tokenString string, expected TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != expected {
		return nil, ErrWrongType
	}
	return claims, nil
}
