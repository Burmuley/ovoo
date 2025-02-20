package entities

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

// Id represents a unique identifier based on ULID (Universally Unique Lexicographically Sortable Identifier)
type Id string

// NewId generates a new Id using ULID
func NewId() Id {
	return Id(ulid.Make().String())
}

// Validate checks if the Id is valid
func (id Id) Validate() error {
	if len(id) == 0 {
		return fmt.Errorf("%w: validating id: can not be empty", ErrValidation)
	}
	if _, err := ulid.Parse(string(id)); err != nil {
		return fmt.Errorf("%w: validating id: %w", ErrValidation, err)
	}

	return nil
}

// String returns the string representation of the Id
func (id Id) String() string {
	return string(id)
}
