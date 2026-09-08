package entities

import (
	"fmt"
	"strings"
	"time"
)

const ApiTokenPrefix = "ovtk"

type ApiTokenBulkUpdateFields struct {
	Description *string
	Active      *bool
	UpdatedById *Id
}

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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   User
}

// NewApiToken creates a new ApiToken with the given expiration, description, and owner.
func NewApiToken(expiration time.Time, name, description string, owner User) (*ApiToken, error) {
	rawToken, err := RandString(32)
	if err != nil {
		return nil, err
	}
	salt, err := RandString(16)
	if err != nil {
		return nil, err
	}
	hash := HashSaltToken(salt, rawToken)
	id := NewId()
	token := &ApiToken{
		ID:          id,
		Name:        name,
		Description: description,
		Owner:       owner,
		Expiration:  expiration,
		Active:      true,
		TokenHash:   hash,
		Token:       strings.Join([]string{ApiTokenPrefix, Base62Encode([]byte(id)), rawToken}, "_"),
		Salt:        salt,
	}

	return token, nil
}

// Validate checks if the ApiToken's fields are valid.
func (t *ApiToken) Validate() error {
	if err := t.ID.Validate(); err != nil {
		return err
	}

	if err := t.Owner.Validate(); err != nil {
		return fmt.Errorf("validating token owner: %w", err)
	}

	if len(t.Name) == 0 {
		return fmt.Errorf("token name can not be empty")
	}

	if len(t.TokenHash) == 0 {
		return fmt.Errorf("token value can not be empty")
	}

	return nil
}

// Expired checks if the ApiToken has expired.
func (t *ApiToken) Expired() bool {
	return time.Now().Compare(t.Expiration) >= 0
}
