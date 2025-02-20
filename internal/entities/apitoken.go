package entities

import (
	"crypto/rand"
	"fmt"
	"time"
)

// ApiToken represents an API token with its associated metadata.
type ApiToken struct {
	ID          Id
	Token       string
	Description string
	Owner       User
	Expiration  time.Time
	Active      bool
}

// NewToken creates a new ApiToken with the given expiration, description, and owner.
func NewToken(expiration time.Time, description string, owner User) *ApiToken {
	return &ApiToken{
		ID:          NewId(),
		Description: description,
		Owner:       owner,
		Expiration:  expiration,
		Active:      true,
		Token:       generateTokenString(),
	}
}

// Validate checks if the ApiToken's fields are valid.
func (t *ApiToken) Validate() error {
	if err := t.ID.Validate(); err != nil {
		return fmt.Errorf("%w: validating token id: %w", ErrValidation, err)
	}

	if err := t.Owner.Validate(); err != nil {
		return fmt.Errorf("%w: validating token owner: %w", ErrValidation, err)
	}

	if len(t.Token) == 0 {
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

// generateTokenString creates a new random token string.
func generateTokenString() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
