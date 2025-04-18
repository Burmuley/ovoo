package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type ApiTokensService struct {
	repoFactory *factory.RepoFactory
}

// NewApiTokensService creates a new ApiTokensService instance.
func NewApiTokensService(repoFabric *factory.RepoFactory) (*ApiTokensService, error) {
	if repoFabric == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &ApiTokensService{repoFactory: repoFabric}, nil
}

// GetById retrieves an API token by ID without validating ownership.
// This should only be used in trusted contexts.
func (t *ApiTokensService) GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error) {
	if err := tokenId.Validate(); err != nil {
		return entities.ApiToken{}, fmt.Errorf("getting api token by id: %w", err)
	}

	token, err := t.repoFactory.ApiTokens.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}

	return token, nil
}

// GetByIdValidOwner retrieves an API token by ID and validates ownership against the provided owner ID.
// Returns an error if token validation fails or if the token doesn't belong to the owner.
func (t *ApiTokensService) GetByIdValidOwner(ctx context.Context, tokenId, ownerId entities.Id) (entities.ApiToken, error) {
	token, err := t.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}

	if token.Owner.ID != ownerId {
		return entities.ApiToken{}, errors.New("token does not belong to the current user")
	}

	return token, nil
}

// GetAll retrieves all API tokens belonging to the specified owner.
// Returns an error if owner ID validation fails.
func (t *ApiTokensService) GetAll(ctx context.Context, ownerId entities.Id) ([]entities.ApiToken, error) {
	if err := ownerId.Validate(); err != nil {
		return nil, fmt.Errorf("getting api tokens for user: %w", err)
	}

	tokens, err := t.repoFactory.ApiTokens.GetAllForUser(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Create generates a new API token for the specified owner with the given name, description,
// and expiration duration (in days). Returns the created token or an error if validation
// or token creation fails.
func (t *ApiTokensService) Create(ctx context.Context, owner entities.User, name, description string, expireIn int) (entities.ApiToken, error) {
	if err := owner.Validate(); err != nil {
		return entities.ApiToken{}, fmt.Errorf("validating new token owner: %w", err)
	}

	token, err := entities.NewToken(time.Now().Add(time.Duration(expireIn*24)*time.Hour), name, description, owner)
	if err != nil {
		return entities.ApiToken{}, fmt.Errorf("generating new token: %w", err)
	}

	if err := t.repoFactory.ApiTokens.Create(ctx, *token); err != nil {
		return entities.ApiToken{}, fmt.Errorf("creating new token: %w", err)
	}

	return *token, nil
}

// Update modifies an existing API token with the provided details.
// Not implemented yet.
func (t *ApiTokensService) Update(ctx context.Context, tokenId, name, description string, active bool) (entities.ApiToken, error) {
	panic("implement me!")
}

// Delete removes an API token with the specified ID.
// Not implemented yet.
func (t *ApiTokensService) Delete(ctx context.Context, tokenId string) error {
	panic("implement me!")
}
