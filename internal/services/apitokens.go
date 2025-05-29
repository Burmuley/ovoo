package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type ApiTokensService struct {
	repof *factory.RepoFactory
}

// NewApiTokensService creates a new ApiTokensService instance.
func NewApiTokensService(repoFabric *factory.RepoFactory) (*ApiTokensService, error) {
	if repoFabric == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	return &ApiTokensService{repof: repoFabric}, nil
}

// GetById retrieves an API token by ID without validating ownership.
// This should only be used in trusted contexts.
func (t *ApiTokensService) GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error) {
	token, err := t.repof.ApiTokens.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}

	return token, nil
}

// GetByIdCurUser retrieves an API token by ID and validates ownership against the provided owner ID.
// Returns an error if token validation fails or if the token doesn't belong to the owner.
func (t *ApiTokensService) GetByIdCurUser(ctx context.Context, cuser entities.User, tokenId entities.Id) (entities.ApiToken, error) {
	token, err := t.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}

	if !canGetApiToken(cuser, token) {
		return entities.ApiToken{}, entities.ErrNotAuthorized
	}

	return token, nil
}

// GetAll retrieves all API tokens belonging to the specified owner.
// Returns an error if owner ID validation fails.
func (t *ApiTokensService) GetAll(ctx context.Context, cuser entities.User) ([]entities.ApiToken, error) {
	tokens, err := t.repof.ApiTokens.GetAllForUser(ctx, cuser.ID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Create generates a new API token for the specified owner with the given name, description,
// and expiration duration (in days). Returns the created token or an error if validation
// or token creation fails.
func (t *ApiTokensService) Create(ctx context.Context, cuser entities.User, name, description string, expireIn int) (entities.ApiToken, error) {
	if !canCreateApiToken(cuser) {
		return entities.ApiToken{}, entities.ErrNotAuthorized
	}

	if name == "" {
		return entities.ApiToken{}, fmt.Errorf("%w: name field cannot be empty", entities.ErrValidation)
	}

	if expireIn < 1 {
		return entities.ApiToken{}, fmt.Errorf("%w: expire_in value cannot be less than 1", entities.ErrValidation)
	}
	token, err := entities.NewToken(time.Now().Add(time.Duration(expireIn*24)*time.Hour), name, description, cuser)
	if err != nil {
		return entities.ApiToken{}, fmt.Errorf("%w: %w", entities.ErrGeneral, err)
	}

	token.UpdatedBy = cuser
	if err := t.repof.ApiTokens.Create(ctx, *token); err != nil {
		return entities.ApiToken{}, err
	}

	return *token, nil
}

// Update modifies an existing API token with the provided details.
// Updates token name, description, and/or active status based on the provided non-nil values.
// Returns an error if trying to activate an expired token.
func (t *ApiTokensService) Update(ctx context.Context, cuser entities.User, tokenId entities.Id, name, description *string, active *bool) (entities.ApiToken, error) {
	token, err := t.repof.ApiTokens.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}

	if !canUpdateApiToken(cuser, token) {
		return entities.ApiToken{}, entities.ErrNotAuthorized
	}

	if name != nil {
		token.Name = *name
	}

	if description != nil {
		token.Description = *description
	}

	if active != nil {
		if *active == true && token.Expired() {
			return entities.ApiToken{}, fmt.Errorf("%w: can not activate expired token", entities.ErrValidation)
		}
		token.Active = *active
	}

	token.UpdatedBy = cuser
	token, err = t.repof.ApiTokens.Update(ctx, token)
	if err != nil {
		return entities.ApiToken{}, err
	}

	return token, nil
}

// Delete removes an API token with the specified ID.
// Permanently removes the token from the repository.
func (t *ApiTokensService) Delete(ctx context.Context, cuser entities.User, tokenId entities.Id) error {
	token, err := t.repof.ApiTokens.GetById(ctx, tokenId)
	if err != nil {
		return err
	}

	if !canDeleteApiToken(cuser, token) {
		return entities.ErrNotAuthorized
	}

	token.UpdatedBy = cuser
	return t.repof.ApiTokens.Delete(ctx, tokenId)
}
