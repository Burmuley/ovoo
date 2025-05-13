package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// AliasesService handles operations related to alias addresses.
type AliasesService struct {
	repoFactory     *factory.RepoFactory
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

	return &AliasesService{repoFactory: repoFabric, domain: domain, wordsDictionary: wordsDict}, nil
}

// Create generates a new alias address and stores it.
func (als *AliasesService) Create(
	ctx context.Context,
	cuser entities.User,
	protAddr entities.Address,
	metadata entities.AddressMetadata,
) (entities.Address, error) {
	if !canCreateAlias(cuser) {
		return entities.Address{}, fmt.Errorf("creating alias: %w", entities.ErrNotAuthorized)
	}

	if err := protAddr.Validate(); err != nil {
		return entities.Address{}, err
	}

	aliasEmail, err := entities.GenAliasEmail(als.domain, als.wordsDictionary)
	if err != nil {
		return entities.Address{}, fmt.Errorf("creating alias: %w", err)
	}

	alias := entities.Address{
		Type:           entities.AliasAddress,
		ID:             entities.NewId(),
		Email:          aliasEmail,
		ForwardAddress: &protAddr,
		Metadata:       metadata,
		Owner:          cuser,
	}

	if err := alias.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("creating alias: %w", err)
	}

	if err := als.repoFactory.Address.Create(ctx, alias); err != nil {
		return entities.Address{}, fmt.Errorf("creating alias: %w", err)
	}

	return alias, nil
}

// Update modifies an existing alias address.
func (als *AliasesService) Update(ctx context.Context, cuser entities.User, alias entities.Address) (entities.Address, error) {
	if !canUpdateAlias(cuser, alias) {
		return entities.Address{}, fmt.Errorf("updating alias: %w", entities.ErrNotAuthorized)
	}

	if err := alias.Validate(); err != nil {
		return entities.Address{}, err
	}

	cur, err := als.repoFactory.Address.GetById(ctx, alias.ID)
	if err != nil {
		return entities.Address{}, fmt.Errorf("updating alias: %w", err)
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

	if err := als.repoFactory.Address.Update(ctx, alias); err != nil {
		return entities.Address{}, fmt.Errorf("updating alias: %w", err)
	}

	return alias, nil
}

// GetAll retrieves all alias addresses for a given owner.
func (als *AliasesService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.Address, error) {
	if filters == nil {
		filters = make(map[string][]string)
	}

	// reset owner value if user is `regular`
	if cuser.Type == entities.RegularUser {
		filters["owner"] = []string{string(cuser.ID)}
	}

	filters["type"] = []string{strconv.Itoa(entities.AliasAddress)}
	aliases, err := als.repoFactory.Address.GetAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("getting aliases: %w", err)
	}

	if cuser.Type == entities.AdminUser {
		return aliases, nil
	}

	fAliases := make([]entities.Address, 0, len(aliases)) // filtered aliases
	for _, alias := range aliases {
		if alias.Owner.ID == cuser.ID {
			fAliases = append(fAliases, alias)
		}
	}

	return fAliases, nil
}

// GetById retrieves an alias address by its ID.
func (als *AliasesService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.Address, error) {
	if err := id.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("getting alias by id: %w", err)
	}

	alias, err := als.repoFactory.Address.GetById(ctx, id)
	if err != nil {
		return entities.Address{}, fmt.Errorf("getting alias by id: %w", err)
	}

	if !canGetAlias(cuser, alias) {
		return entities.Address{}, fmt.Errorf("getting alias by id: %w", entities.ErrNotAuthorized)
	}

	return alias, nil
}

func (als *AliasesService) DeleteById(ctx context.Context, cuser entities.User, id entities.Id) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("deleting alias by id: %w", err)
	}

	alias, err := als.repoFactory.Address.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting alias by id: %w", err)
	}

	if !canDeleteAlias(cuser, alias) {
		return fmt.Errorf("deleting alias by id: %w", entities.ErrNotAuthorized)
	}

	if err := als.repoFactory.Address.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("deleting alias by id: %w", err)
	}

	return nil
}
