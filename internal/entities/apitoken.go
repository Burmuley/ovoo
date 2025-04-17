package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const ApiTokenPrefix = "ovtk"

// ApiToken represents an API token with its associated metadata.
type ApiToken struct {
	ID          Id
	Name        string
	Token       string // for runtime purposes only, should not be stored
	TokenHash   string
	Salt        string
	Description string
	Owner       User
	Expiration  time.Time
	Active      bool
}

// NewToken creates a new ApiToken with the given expiration, description, and owner.
func NewToken(expiration time.Time, name, description string, owner User) (*ApiToken, error) {
	rawToken, err := RandString(32)
	if err != nil {
		return nil, err
	}
	salt, err := RandString(16)
	if err != nil {
		return nil, err
	}
	hash := HashApiToken(salt, rawToken)
	id := NewId()
	token := &ApiToken{
		ID:          id,
		Name:        name,
		Description: description,
		Owner:       owner,
		Expiration:  expiration,
		Active:      true,
		TokenHash:   hash,
		Token:       strings.Join([]string{ApiTokenPrefix, string(id), rawToken}, "_"),
		Salt:        salt,
	}

	return token, nil
}

// Validate checks if the ApiToken's fields are valid.
func (t *ApiToken) Validate() error {
	if err := t.ID.Validate(); err != nil {
		return fmt.Errorf("%w: validating token id: %w", ErrValidation, err)
	}

	if err := t.Owner.Validate(); err != nil {
		return fmt.Errorf("%w: validating token owner: %w", ErrValidation, err)
	}

	if len(t.TokenHash) == 0 {
		return fmt.Errorf("%w: validating token: token value can not be empty", ErrValidation)
	}

	return nil
}

// Expired checks if the ApiToken has expired.
func (t *ApiToken) Expired() bool {
	if time.Now().Compare(t.Expiration) >= 0 {
		return true
	}

	return false
}

// HashApiToken creates a SHA-256 hash of a token by prepending the salt.
// The salt adds randomness to prevent rainbow table attacks.
// Returns the hex-encoded hash string.
func HashApiToken(salt, token string) string {
	sum := sha256.Sum256(append([]byte(salt), []byte(token)...))
	return hex.EncodeToString(sum[:])
}
