package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type DomainCreateCmd struct {
	Name   string
	Global bool
}

type DomainUpdateCmd struct {
	DomainId entities.Id
	Active   *bool
}

type DomainsService struct {
	repof *factory.RepoFactory
}

func NewDomainsService(repoFabric *factory.RepoFactory) (*DomainsService, error) {
	if repoFabric == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &DomainsService{repof: repoFabric}, nil
}

func (d *DomainsService) GetAll(ctx context.Context, cuser entities.User) ([]entities.CustomDomain, error) {
	var filter entities.CustomDomainFilter

	if cuser.Type != entities.AdminUser {
		filter = entities.CustomDomainFilter{
			Owners:        []entities.Id{cuser.ID},
			IncludeGlobal: true,
		}
	}

	domains, _, err := d.repof.Domain.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (d *DomainsService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.CustomDomain, error) {
	domain, err := d.repof.Domain.GetById(ctx, id)
	if err != nil {
		return entities.CustomDomain{}, err
	}

	if !canUpdateDomain(cuser, domain) {
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

	now := time.Now()
	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      strings.TrimSpace(cmd.Name),
		Global:    cmd.Global,
		Owner:     cuser,
		Active:    true,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
		UpdatedBy: cuser,
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
