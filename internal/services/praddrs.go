package services

import (
	"context"
	"fmt"
	"slices"
	"strconv"

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
func (prs *ProtectedAddrService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.Address, error) {
	if filters == nil {
		filters = make(map[string][]string)
	}

	// by default all requests with no 'owner' filter set are limited to the current user
	if len(filters["owner"]) == 0 {
		filters["owner"] = []string{string(cuser.ID)}
	} else {
		// if 'owner' filter has entry 'all' - remove filter to retrieve all entries for admin user
		if cuser.Type == entities.AdminUser && slices.Contains(filters["owner"], "all") {
			delete(filters, "owner")
		}

		// reset 'owner' filter to the current user for non-admins
		if cuser.Type != entities.AdminUser {
			filters["owner"] = []string{string(cuser.ID)}
		}
	}

	filters["type"] = []string{strconv.Itoa(entities.ProtectedAddress)}
	praddrs, err := prs.repof.Address.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return praddrs, nil
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

	if err := prs.repof.Address.DeleteById(ctx, id); err != nil {
		return err
	}

	return nil
}
