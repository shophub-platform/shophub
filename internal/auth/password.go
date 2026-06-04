// Package auth pruža heširanje lozinki i JWT izdavanje/validaciju.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword vraća bcrypt heš date lozinke.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword proverava da li lozinka odgovara sačuvanom bcrypt hešu.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
