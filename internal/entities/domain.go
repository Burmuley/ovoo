package entities

import (
	"fmt"
	"time"
)

type CustomDomain struct {
	ID                Id
	Name              string
	Global            bool
	Owner             User
	CreatedAt         time.Time
	UpdatedAt         time.Time
	UpdatedBy         User
	Active            bool
	Verified          bool
	VerifiedAt        time.Time
	VerificationToken string
}

func (cd *CustomDomain) Validate() error {
	if err := cd.ID.Validate(); err != nil {
		return err
	}

	if err := cd.Owner.Validate(); err != nil {
		return fmt.Errorf("validating domain owner: %w", err)
	}

	return nil
}
