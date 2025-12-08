package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type AliasCreateCmd struct {
	ProtectedAddressId string
	Metadata           struct {
		Comment     *string
		ServiceName *string
	}
}

type AliasUpdateCmd struct {
	AliasId  entities.Id
	Metadata struct {
		Comment     *string
		ServiceName *string
	}
}

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
	cmd AliasCreateCmd,
	// protAddrId entities.Id,
	// metadata entities.AddressMetadata,
) (entities.Address, error) {
	if !canCreateAlias(cuser) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	protAddr, err := als.repof.Address.GetById(ctx, entities.Id(cmd.ProtectedAddressId))
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
	}

	if err := protAddr.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	aliasEmail, err := entities.GenAliasEmail(als.domain, als.wordsDictionary)
	if err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrGeneral, err)
	}

	metadata := entities.AddressMetadata{}
	if cmd.Metadata.ServiceName != nil {
		metadata.ServiceName = strings.TrimSpace(*cmd.Metadata.ServiceName)
	}

	if cmd.Metadata.Comment != nil {
		metadata.Comment = strings.TrimSpace(*cmd.Metadata.Comment)
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
func (als *AliasesService) Update(ctx context.Context, cuser entities.User, cmd AliasUpdateCmd) (entities.Address, error) {
	alias, err := als.repof.Address.GetById(ctx, cmd.AliasId)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
	}

	if !canUpdateAlias(cuser, alias) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	alias.UpdatedBy = cuser
	if cmd.Metadata.Comment != nil {
		alias.Metadata.Comment = strings.TrimSpace(*cmd.Metadata.Comment)
	}

	if cmd.Metadata.ServiceName != nil {
		alias.Metadata.ServiceName = strings.TrimSpace(*cmd.Metadata.ServiceName)
	}

	if err := alias.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := als.repof.Address.Update(ctx, alias); err != nil {
		return entities.Address{}, err
	}

	return alias, nil
}

// GetAll retrieves all alias addresses for a given owner.
func (als *AliasesService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.Address, entities.PaginationMetadata, error) {
	if !canGetAliases(cuser) {
		return []entities.Address{}, entities.PaginationMetadata{}, entities.ErrNotAuthorized
	}

	filter, err := entities.NewAddressFilter(filters)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}
	filter.Types = []entities.AddressType{entities.AliasAddress}

	// reset Owners filter for non-admins
	if cuser.Type != entities.AdminUser {
		filter.Owners = []entities.Id{cuser.ID}
	} else if slices.Contains(filter.Owners, "all") && cuser.Type == entities.AdminUser {
		filter.Owners = nil
	} else if filter.Owners == nil {
		filter.Owners = []entities.Id{cuser.ID}
	}

	return als.repof.Address.GetAll(ctx, filter)
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

	if err := deleteAliasIds(ctx, als.repof, []entities.Id{alias.ID}); err != nil {
		return err
	}

	return nil
}
