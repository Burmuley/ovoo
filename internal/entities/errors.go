package entities

import "errors"

var (
	// ErrValidation is returned when there's a validation error
	ErrValidation = errors.New("validation error")
	// ErrDuplicateEntry is returned when a duplicate entry is detected
	ErrDuplicateEntry = errors.New("duplicate entry")
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("not found")
	// ErrConfiguration is returned when there's a configuration error
	ErrConfiguration = errors.New("configuration error")
	// ErrGeneral is returned for general backend errors
	ErrGeneral = errors.New("general backend error")
	// ErrNotAuthorized is returned when operation requested by the user is not authorized
	ErrNotAuthorized = errors.New("requested operation is not authorized for the user")
	// ErrDatabase is returned when database operation failed, except when "not found" is returned
	ErrDatabase = errors.New("database error")
)
