package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTM() *TokenManager {
	return NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
}

func TestAccessTokenRoundTrip(t *testing.T) {
	tm := newTM()
	uid := uuid.New()
	tok, err := tm.NewAccessToken(uid, "a@b.com")
	if err != nil {
		t.Fatalf("izdavanje: %v", err)
	}
	claims, err := tm.Parse(tok, AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != uid || claims.Email != "a@b.com" || claims.Type != AccessToken {
		t.Fatalf("neočekivani claims: %+v", claims)
	}
}

func TestRefreshTokenVersionPreserved(t *testing.T) {
	tm := newTM()
	tok, _ := tm.NewRefreshToken(uuid.New(), "a@b.com", 7)
	claims, err := tm.Parse(tok, RefreshToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Version != 7 {
		t.Fatalf("Version = %d, očekivano 7", claims.Version)
	}
}

func TestWrongTypeRejected(t *testing.T) {
	tm := newTM()
	tok, _ := tm.NewAccessToken(uuid.New(), "a@b.com")
	if _, err := tm.Parse(tok, RefreshToken); err == nil {
		t.Fatal("očekivana greška za pogrešan tip tokena")
	}
}

func TestBadSecretRejected(t *testing.T) {
	tok, _ := newTM().NewAccessToken(uuid.New(), "a@b.com")
	other := NewTokenManager("drugi-secret", time.Minute, time.Hour)
	if _, err := other.Parse(tok, AccessToken); err == nil {
		t.Fatal("očekivana greška za token potpisan drugim secret-om")
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	tm := NewTokenManager("s", -time.Minute, -time.Minute)
	tok, _ := tm.NewAccessToken(uuid.New(), "a@b.com")
	if _, err := tm.Parse(tok, AccessToken); err == nil {
		t.Fatal("očekivana greška za istekao token")
	}
}

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("heš: %v", err)
	}
	if !CheckPassword(hash, "supersecret") {
		t.Fatal("ispravna lozinka odbijena")
	}
	if CheckPassword(hash, "pogresna") {
		t.Fatal("pogrešna lozinka prihvaćena")
	}
}
