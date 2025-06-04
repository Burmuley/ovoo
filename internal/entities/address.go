package entities

import (
	"fmt"
	"time"
)

type AddressType int8

const (
	AliasAddress AddressType = iota
	ReplyAliasAddress
	ProtectedAddress
	ExternalAddress
)

// AddressMetadata contains additional information about an address.
// It includes a comment for user notes and the name of the associated service.
type AddressMetadata struct {
	Comment     string
	ServiceName string
}

// Address represents an email address with associated metadata and ownership information.
// It can be of different types (alias, reply alias, protected, or external) and may have
// a forward address for routing purposes.
type Address struct {
	ID             Id
	Type           AddressType
	Email          Email
	ForwardAddress *Address
	Owner          User
	Metadata       AddressMetadata
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UpdatedBy      User
}

// Validate checks if the Address object is valid according to the defined rules.
// It returns an error if any validation fails, or nil if the Address is valid.
// The validation includes checking the ID, email, forward address (if applicable),
// and owner ID. It also enforces rules specific to protected and external addresses.
func (a *Address) Validate() error {
	if err := a.ID.Validate(); err != nil {
		return err
	}
	// protected address can not have ForwardEmail set
	if a.Type == ProtectedAddress && a.ForwardAddress != nil {
		return fmt.Errorf("protected address can not have forward email set")
	}

	// external address can not have ForwardEmail set
	if a.Type == ExternalAddress && a.ForwardAddress != nil {
		return fmt.Errorf("external address can not have forward email set")
	}

	// Emails should be valid email
	if err := a.Email.Validate(); err != nil {
		return fmt.Errorf("validating address email: %w", err)
	}

	if a.Type != ProtectedAddress && a.Type != ExternalAddress {
		if err := a.ForwardAddress.Validate(); err != nil {
			return fmt.Errorf("validating address forward email: %w", err)
		}
	}

	// owner should be set
	if err := a.Owner.ID.Validate(); err != nil {
		return fmt.Errorf("validating owner: %w", err)
	}

	return nil
}
