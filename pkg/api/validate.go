package api

import (
	"github.com/oklog/ulid/v2"
	"net/mail"
)

func validateEmail(e string) bool {
	_, err := mail.ParseAddress(e)
	return err == nil
}

func validateId(id string) bool {
	_, err := ulid.Parse(id)
	return err == nil
}
