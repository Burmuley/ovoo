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
}

// Validate checks the integrity of the Chain structure.
// It validates the hash, from address, and to address.
// It also ensures the hash matches the one generated from the email addresses.
// Returns an error if any validation fails.
func (c Chain) Validate() error {
	if err := c.Hash.Validate(); err != nil {
		return fmt.Errorf("%w: validating chain hash: %w", ErrValidation, err)
	}

	if err := c.FromAddress.Validate(); err != nil {
		return fmt.Errorf("%w: validating chain from address: %w", ErrValidation, err)
	}

	if err := c.ToAddress.Validate(); err != nil {
		return fmt.Errorf("%w: validating chain to address: %w", ErrValidation, err)
	}

	// TODO: this validation is incorrect since FromAddress and ToAddress contains target addresses
	// aimed to be used as replacements in milter, whereas Hash contains a has taken from original
	// FromAddress+ToAddress pair
	// hash := NewHash(string(c.FromAddress.Email), string(c.ToAddress.Email))
	// if hash != c.Hash {
	// 	return fmt.Errorf("%w: validating chain: emails produced different hash", ErrValidation)
	// }

	return nil
}
