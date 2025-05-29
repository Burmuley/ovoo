package entities

import (
	"fmt"
	"time"
)

// Chain represents a link between two addresses with a hash and creation timestamp.
type Chain struct {
	Hash            Hash
	FromAddress     Address
	ToAddress       Address
	OrigFromAddress Address
	OrigToAddress   Address
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UpdatedBy       User
}

// Validate checks the integrity of the Chain structure.
// It validates the hash, from address, and to address.
// It also ensures the hash matches the one generated from the email addresses.
// Returns an error if any validation fails.
func (c Chain) Validate() error {
	if err := c.Hash.Validate(); err != nil {
		return err
	}

	if err := c.FromAddress.Validate(); err != nil {
		return fmt.Errorf("validating chain from address: %w", err)
	}

	if err := c.ToAddress.Validate(); err != nil {
		return fmt.Errorf("validating chain to address: %w", err)
	}

	hash := NewHash(string(c.OrigFromAddress.Email), string(c.OrigToAddress.Email))
	if hash != c.Hash {
		return fmt.Errorf("emails produced different hash")
	}

	return nil
}
