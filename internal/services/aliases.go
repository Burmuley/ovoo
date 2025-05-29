package services

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// AliasesService handles operations related to alias addresses.
type AliasesService struct {
	repof           *factory.RepoFactory
	domain          string
	wordsDictionary []string
}

// NewAliasesService creates a new AliasesUsecase instance.
func NewAliasesService(domain string, wordsDict []string, repoFabric *factory.RepoFactory) (*AliasesService, error) {
	if len(domain) < 2 {
		return nil, fmt.Errorf("%w: invalid domain defined", entities.ErrConfiguration)
	}

	if len(wordsDict) == 0 {
		return nil, fmt.Errorf("%w: words dictionary can not be empty", entities.ErrConfiguration)
	}

	if repoFabric == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &AliasesService{repof: repoFabric, domain: domain, wordsDictionary: wordsDict}, nil
}

// Create generates a new alias address and stores it.
func (als *AliasesService) Create(
	ctx context.Context,
	cuser entities.User,
	protAddr entities.Address,
	metadata entities.AddressMetadata,
) (entities.Address, error) {
	if !canCreateAlias(cuser) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	if err := protAddr.Validate(); err != nil {
		return entities.Address{}, err
	}

	aliasEmail, err := entities.GenAliasEmail(als.domain, als.wordsDictionary)
	if err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrGeneral, err)
	}

	alias := entities.Address{
		Type:           entities.AliasAddress,
		ID:             entities.NewId(),
		Email:          aliasEmail,
		ForwardAddress: &protAddr,
		Metadata:       metadata,
		Owner:          cuser,
		UpdatedBy:      cuser,
	}

	if err := alias.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := als.repof.Address.Create(ctx, alias); err != nil {
		return entities.Address{}, err
	}

	return alias, nil
}

// Update modifies an existing alias address.
func (als *AliasesService) Update(ctx context.Context, cuser entities.User, alias entities.Address) (entities.Address, error) {
	if !canUpdateAlias(cuser, alias) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	if err := alias.Validate(); err != nil {
		return entities.Address{}, err
	}

	cur, err := als.repof.Address.GetById(ctx, alias.ID)
	if err != nil {
		return entities.Address{}, err
	}

	// validate fields
	if cur.Type != alias.Type {
		return entities.Address{}, fmt.Errorf("%w: address type can not be changed", entities.ErrValidation)
	}

	if cur.Email != alias.Email {
		return entities.Address{}, fmt.Errorf("%w: alias email can not be changed", entities.ErrValidation)
	}

	if cur.ForwardAddress != alias.ForwardAddress {
		return entities.Address{}, fmt.Errorf(
			"%w: forward address can not be changed for alias address",
			entities.ErrValidation,
		)
	}

	if cur.Owner != alias.Owner {
		return entities.Address{}, fmt.Errorf("%w: alias owner can not be changed", entities.ErrValidation)
	}

	alias.UpdatedBy = cuser
	if err := als.repof.Address.Update(ctx, alias); err != nil {
		return entities.Address{}, err
	}

	return alias, nil
}

// GetAll retrieves all alias addresses for a given owner.
func (als *AliasesService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.Address, error) {
	if !canGetAliases(cuser) {
		return []entities.Address{}, entities.ErrNotAuthorized
	}

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

	filters["type"] = []string{strconv.Itoa(entities.AliasAddress)}
	aliases, err := als.repof.Address.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return aliases, nil
}

// GetById retrieves an alias address by its ID.
func (als *AliasesService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.Address, error) {
	if err := id.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	alias, err := als.repof.Address.GetById(ctx, id)
	if err != nil {
		return entities.Address{}, err
	}

	if !canGetAlias(cuser, alias) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	return alias, nil
}

func (als *AliasesService) DeleteById(ctx context.Context, cuser entities.User, id entities.Id) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	alias, err := als.repof.Address.GetById(ctx, id)
	if err != nil {
		return err
	}

	if !canDeleteAlias(cuser, alias) {
		return entities.ErrNotAuthorized
	}

	alias.UpdatedBy = cuser
	if err := als.repof.Address.DeleteById(ctx, id); err != nil {
		return err
	}

	return nil
}
