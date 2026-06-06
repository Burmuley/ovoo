package entities

import (
	"fmt"
	"strings"
	"time"
)

type DNSRecordType string

const (
	CNAMERecord DNSRecordType = "cname"
	TXTRecord   DNSRecordType = "txt"
)

type DomainVerificationData struct {
	RecordType             DNSRecordType
	Name                   string
	Value                  string
	LastVerificationResult string
}

type CustomDomain struct {
	ID               Id
	Name             string
	Global           bool
	Owner            User
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UpdatedBy        User
	Active           bool
	Verified         bool
	VerifiedAt       time.Time
	VerificationData DomainVerificationData
}

func (cd *CustomDomain) Validate() error {
	if err := cd.ID.Validate(); err != nil {
		return err
	}

	if len(strings.TrimSpace(cd.Name)) == 0 {
		return fmt.Errorf("validating domain name: can not be empty")
	}

	if err := cd.Owner.Validate(); err != nil {
		return fmt.Errorf("validating domain owner: %w", err)
	}

	return nil
}
