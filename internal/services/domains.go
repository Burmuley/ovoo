package services

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type DomainCreateCmd struct {
	Name                   string
	Global                 bool
	VerificationRecordType string
}

type DomainVerifyCmd struct {
	DomainId entities.Id
}

type DomainUpdateCmd struct {
	DomainId entities.Id
	Active   *bool
}

const (
	domainVerifyRecordPrefix     string = "_ovoo_check_"
	domainVerifyTXTValuePrefix   string = "OVOO_ID="
	domainVerifyCNAMEValueSuffix        = ".ovoocheck.local."
)

type DomainsService struct {
	repof *factory.RepoFactory
}

func NewDomainsService(repoFabric *factory.RepoFactory) (*DomainsService, error) {
	if repoFabric == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &DomainsService{repof: repoFabric}, nil
}

func (d *DomainsService) GetAll(ctx context.Context, cuser entities.User, filters entities.CustomDomainFilter) ([]entities.CustomDomain, entities.PaginationMetadata, error) {
	if !canGetDomains(cuser) {
		return nil, entities.PaginationMetadata{}, entities.ErrNotAuthorized
	}

	// admin and milter users can read all domains in the system
	// regular users are limited to only read domains they own
	if cuser.Type != entities.AdminUser && cuser.Type != entities.MilterUser {
		filters.Owners = []entities.Id{cuser.ID}
	}

	// milter can only receive active&verified domains + global
	if cuser.Type == entities.MilterUser {
		filters.Active = new(true)
		filters.Verified = new(true)
		filters.IncludeGlobal = true
	}

	domains, pgm, err := d.repof.Domain.GetAll(ctx, filters)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}

	return domains, pgm, nil
}

func (d *DomainsService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.CustomDomain, error) {
	domain, err := d.repof.Domain.GetById(ctx, id)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	if !canGetDomain(cuser, domain) {
		return entities.CustomDomain{}, entities.ErrNotAuthorized
	}

	return domain, nil
}

func (d *DomainsService) Create(ctx context.Context, cuser entities.User, cmd DomainCreateCmd) (entities.CustomDomain, error) {
	if !canCreateDomain(cuser) {
		return entities.CustomDomain{}, entities.ErrNotAuthorized
	}

	if cmd.Global && !canCreateGlobalDomain(cuser) {
		return entities.CustomDomain{}, entities.ErrNotAuthorized
	}

	if strings.TrimSpace(cmd.Name) == "" {
		return entities.CustomDomain{}, fmt.Errorf("%w: name field cannot be empty", entities.ErrValidation)
	}

	if domain, err := d.repof.Domain.GetByName(ctx, cmd.Name); err == nil && domain.ID.Validate() == nil {
		return entities.CustomDomain{}, fmt.Errorf("%w: domain already exists", entities.ErrDuplicateEntry)
	}

	now := time.Now()
	vd, err := fillVerificationData(
		entities.DomainVerificationData{
			RecordType: entities.DNSRecordType(cmd.VerificationRecordType),
		},
	)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	domain := entities.CustomDomain{
		ID:               entities.NewId(),
		Name:             strings.TrimSpace(cmd.Name),
		Global:           cmd.Global,
		Owner:            cuser,
		Active:           true,
		Verified:         false,
		CreatedAt:        now,
		UpdatedAt:        now,
		UpdatedBy:        cuser,
		VerificationData: vd,
	}

	if err := domain.Validate(); err != nil {
		return entities.CustomDomain{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := d.repof.Domain.Create(ctx, domain); err != nil {
		return entities.CustomDomain{}, err
	}

	return domain, nil
}

func (d *DomainsService) Update(ctx context.Context, cuser entities.User, cmd DomainUpdateCmd) (entities.CustomDomain, error) {
	domain, err := d.repof.Domain.GetById(ctx, cmd.DomainId)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	if !canUpdateDomain(cuser, domain) {
		return entities.CustomDomain{}, entities.ErrNotAuthorized
	}

	if cmd.Active != nil {
		domain.Active = *cmd.Active
	}

	domain.UpdatedBy = cuser
	domain.UpdatedAt = time.Now()

	if err := domain.Validate(); err != nil {
		return entities.CustomDomain{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	domain, err = d.repof.Domain.Update(ctx, domain)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	return domain, nil
}

func (d *DomainsService) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	domain, err := d.repof.Domain.GetById(ctx, id)
	if err != nil {
		return err
	}

	if !canDeleteDomain(cuser, domain) {
		return entities.ErrNotAuthorized
	}

	return d.repof.Domain.Delete(ctx, cuser, id)
}

func (d *DomainsService) Verify(ctx context.Context, cuser entities.User, id entities.Id) (entities.CustomDomain, error) {
	domain, err := d.repof.Domain.GetById(ctx, id)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	if !canVerifyDomain(cuser, domain) {
		return entities.CustomDomain{}, entities.ErrNotAuthorized
	}

	// if verification failed -> store last error
	if err := verifyDomainDNS(ctx, domain); err != nil {
		domain.VerificationData.LastVerificationResult = err.Error()
	} else {
		// mark as verified in case of success
		domain.Verified = true
		domain.VerifiedAt = time.Now().UTC()
	}

	if _, err := d.repof.Domain.Update(ctx, domain); err != nil {
		return entities.CustomDomain{}, err
	}

	return domain, nil
}

func verifyDomainDNS(ctx context.Context, domain entities.CustomDomain) error {
	targetName := strings.Join([]string{domain.VerificationData.Name, domain.Name}, ".")
	targetValue := domain.VerificationData.Value

	switch domain.VerificationData.RecordType {
	case entities.TXTRecord:
		values, err := net.DefaultResolver.LookupTXT(ctx, targetName)
		if err != nil {
			return fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		if slices.Contains(values, targetValue) {
			return nil
		}

		return fmt.Errorf("%w: invalid target value", entities.ErrValidation)
	case entities.CNAMERecord:
		value, err := net.DefaultResolver.LookupCNAME(ctx, targetName)
		if err != nil {
			return fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		if value == targetValue {
			return nil
		}

		return fmt.Errorf("%w: invalid target value", entities.ErrValidation)
	default:
		return fmt.Errorf(
			"%w: unsupported record type %q",
			entities.ErrValidation,
			string(domain.VerificationData.RecordType),
		)
	}
}
