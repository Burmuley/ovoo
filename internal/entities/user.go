package entities

import (
	"fmt"
	"strings"
)

type UserType int8

const (
	RegularUser UserType = iota
	AdminUser
	MilterUser
)

// User represents a user in the system with various attributes.
type User struct {
	Type         UserType
	ID           Id
	Login        string
	FirstName    string
	LastName     string
	PasswordHash string
}

// Validate checks if the User object is valid and returns an error if not.
func (u User) Validate() error {
	if err := u.ID.Validate(); err != nil {
		return fmt.Errorf("%w: validating user: %w", ErrValidation, err)
	}

	if len(u.Login) == 0 {
		return fmt.Errorf("%w: validating user: login can not be empty", ErrValidation)
	}

	if u.Type > 2 {
		return fmt.Errorf("%w: validating user: invalid user type", ErrValidation)
	}

	return nil
}

// String returns a string representation of the User, combining FirstName and LastName.
func (u User) String() string {
	return strings.TrimSpace(strings.Join([]string{u.FirstName, u.LastName}, " "))
}
