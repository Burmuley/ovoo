package services

import (
	"context"
	"fmt"
	"slices"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// ProtectedAddrService handles operations related to protected addresses
type ProtectedAddrService struct {
	repof *factory.RepoFactory
}

// NewProtectedAddrService creates a new ProtectedAddrUsecase
func NewProtectedAddrService(repoFactory *factory.RepoFactory) (*ProtectedAddrService, error) {
	if repoFactory == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &ProtectedAddrService{repof: repoFactory}, nil
}

// Create creates a new protected address
func (prs *ProtectedAddrService) Create(ctx context.Context, cuser entities.User, protEmail entities.Email, metadata entities.AddressMetadata) (entities.Address, error) {
	if !canCreatePrAddr(cuser) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	if err := protEmail.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	// check if protected address with the email already exists
	if addrs, err := prs.repof.Address.GetByEmail(ctx, protEmail); err == nil {
		for _, addr := range addrs {
			if addr.Type == entities.ProtectedAddress {
				return entities.Address{}, fmt.Errorf("%w: %s", entities.ErrDuplicateEntry, protEmail)
			}
		}
	}

	praddr := entities.Address{
		Type:      entities.ProtectedAddress,
		ID:        entities.NewId(),
		Email:     protEmail,
		Metadata:  metadata,
		Owner:     cuser,
		UpdatedBy: cuser,
	}

	if err := prs.repof.Address.Create(ctx, praddr); err != nil {
		return entities.Address{}, err
	}

	return praddr, nil
}

// Update updates an existing protected address
func (prs *ProtectedAddrService) Update(ctx context.Context, cuser entities.User, praddr entities.Address) (entities.Address, error) {
	if !canUpdatePrAddr(cuser, praddr) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	if err := praddr.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	cur, err := prs.repof.Address.GetById(ctx, praddr.ID)
	if err != nil {
		return entities.Address{}, err
	}

	// validate fields
	if cur.Type != praddr.Type {
		return entities.Address{}, fmt.Errorf("%w: address type can not be changed", entities.ErrValidation)
	}

	if cur.Email != praddr.Email {
		return entities.Address{}, fmt.Errorf("%w: alias email can not be changed", entities.ErrValidation)
	}

	if cur.ForwardAddress != praddr.ForwardAddress {
		return entities.Address{}, fmt.Errorf("%w: forward address can not be changed for alias address", entities.ErrValidation)
	}

	if cur.Owner != praddr.Owner {
		return entities.Address{}, fmt.Errorf("%w: address owner can not be changed", entities.ErrValidation)
	}

	praddr.UpdatedBy = cuser
	if err := prs.repof.Address.Update(ctx, praddr); err != nil {
		return entities.Address{}, err
	}

	return praddr, nil
}

// GetAll retrieves all protected addresses for a given owner
func (prs *ProtectedAddrService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.Address, entities.PaginationMetadata, error) {
	filter, err := entities.NewAddressFilter(filters)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}
	filter.Types = []entities.AddressType{entities.ProtectedAddress}

	// reset Owners filter for non-admins
	if cuser.Type != entities.AdminUser {
		filter.Owners = []entities.Id{cuser.ID}
	} else if slices.Contains(filter.Owners, "all") && cuser.Type == entities.AdminUser {
		filter.Owners = nil
	} else if filter.Owners == nil {
		filter.Owners = []entities.Id{cuser.ID}
	}

	return prs.repof.Address.GetAll(ctx, filter)
}

// GetById retrieves a protected address by its ID
func (prs *ProtectedAddrService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.Address, error) {
	if err := id.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddr, err := prs.repof.Address.GetById(ctx, id)
	if err != nil {
		return entities.Address{}, err
	}

	if !canGetPrAddr(cuser, praddr) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	return praddr, nil
}

func (prs *ProtectedAddrService) GetByEmail(ctx context.Context, cuser entities.User, email entities.Email) (entities.Address, error) {
	if err := email.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddrs, err := prs.repof.Address.GetByEmail(ctx, entities.Email(email))
	if err != nil {
		return entities.Address{}, err
	}

	for _, praddr := range praddrs {
		if praddr.Type == entities.ProtectedAddress {
			if !canGetPrAddr(cuser, praddr) {
				return entities.Address{}, entities.ErrNotAuthorized
			}
			return praddr, nil
		}
	}

	return entities.Address{}, fmt.Errorf("%w: %s", entities.ErrNotFound, email)
}

func (prs *ProtectedAddrService) DeleteById(ctx context.Context, cuser entities.User, id entities.Id) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddr, err := prs.repof.Address.GetById(ctx, id)
	if err != nil {
		return err
	}

	if !canDeletePrAddr(cuser, praddr) {
		return entities.ErrNotAuthorized
	}

	// first - retrieve all aliases and delete them
	if err := deleteAliasesForPrAddr(ctx, prs.repof, praddr); err != nil {
		return err
	}

	// delete protected address after all referencing entities (aliases) has been deleted successfully
	if err := prs.repof.Address.DeleteById(ctx, id); err != nil {
		return err
	}

	return nil
}
