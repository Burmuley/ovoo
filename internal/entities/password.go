package entities

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// NewPasswordHash returns a bcrypt hash of password.
// An empty password is rejected: accounts without a password are represented by an empty
// hash and are authenticated through OIDC only, never by hashing the empty string.
func NewPasswordHash(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("%w: password can not be empty", ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ValidPassword reports whether password matches hashedPassword.
// An empty hash means password authentication is disabled for the account, and an empty
// password never matches - older releases stored a bcrypt hash of the empty string, and
// those accounts must not be reachable with empty credentials.
func ValidPassword(password, hashedPassword string) bool {
	if len(password) == 0 || len(hashedPassword) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
