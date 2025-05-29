package entities

import (
	"net/mail"
)

// Email represents an email address as a string type
type Email string

// Validate checks if the Email is a valid email address
func (e Email) Validate() error {
	_, err := mail.ParseAddress(string(e))
	if err != nil {
		return err
	}
	return nil
}

// String returns the string representation of the Email
func (e Email) String() string {
	return string(e)
}
