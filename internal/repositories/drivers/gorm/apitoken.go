package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/gorm"
)

// TokenGORMRepo implements the TokensReadWriter interface using GORM.
type TokenGORMRepo struct {
	db *gorm.DB
}

// NewApiTokenGORMRepo creates a new TokenGORMRepo instance.
// It returns an error if the provided database connection is nil.
func NewApiTokenGORMRepo(db *gorm.DB) (repositories.TokensReadWriter, error) {
	if db == nil {
		return &TokenGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}

	return &TokenGORMRepo{db: db}, nil
}

// Create adds a new API token to the database.
// It takes a context and an API token entity, converts it to a GORM model,
// and persists it to the database.
// Returns an error if the database operation fails.
func (t *TokenGORMRepo) Create(ctx context.Context, token entities.ApiToken) error {
	gorm_token := apiTokenFromEntity(token)
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Create(&gorm_token).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// BatchCreate adds multiple API tokens to the database in a single operation.
// It takes a context and a slice of API token entities, converts them to GORM models,
// and persists them to the database.
// Returns an error if the database operation fails.
func (t *TokenGORMRepo) BatchCreate(ctx context.Context, tokens []entities.ApiToken) error {
	gorm_tokens := apiTokenFromEntityList(tokens)
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Create(&gorm_tokens).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Delete removes an API token from the database based on its ID.
// It first checks if the token exists by calling GetById.
// If the token exists, it performs a hard delete (unscoped) from the database.
// Returns an error if the token doesn't exist or if the delete operation fails.
func (t *TokenGORMRepo) Delete(ctx context.Context, id entities.Id) error {
	if _, err := t.GetById(ctx, id); err != nil {
		return wrapGormError(err)
	}

	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Unscoped().
		Delete(&ApiToken{}, "id = ?", id.String()).Error; err != nil {
		return wrapGormError(err)
	}
	return nil
}

func (t *TokenGORMRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	// BatchDeleteById removes multiple API tokens from the database based on their IDs.
	// It performs a hard delete (unscoped) for all tokens whose IDs are provided in the 'ids' slice.
	// Returns an error if the delete operation fails.
	if err := t.db.WithContext(ctx).Unscoped().Delete(&ApiToken{}, "id IN ?", ids).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// BatchDeleteForUser removes all API tokens belonging to a specific user from the database.
// It performs a hard delete (unscoped) for all tokens where the 'owner_id' matches the provided user ID.
// Returns an error if the delete operation fails.
func (t *TokenGORMRepo) BatchDeleteForUser(ctx context.Context, id entities.Id) error {
	if err := t.db.WithContext(ctx).Unscoped().Delete(&ApiToken{}, "owner_id = ?", id).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// GetById retrieves an API token from the database based on its ID.
// It also preloads the Owner relation for the token.
// Returns the token as an entity and an error if the token doesn't exist or if the query fails.
func (t *TokenGORMRepo) GetById(ctx context.Context, token_id entities.Id) (entities.ApiToken, error) {
	token := ApiToken{}
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Where("id = ?", token_id).Preload("Owner").First(&token).Error; err != nil {
		return entities.ApiToken{}, wrapGormError(err)
	}

	return apiTokenToEntity(token), nil
}

// GetAllForUser retrieves all API tokens associated with a given user.
// It takes a context and a user ID, finds all tokens associated with that user,
// and preloads the Owner relation for each token.
// Returns a slice of API token entities and an error if the query fails.
func (t *TokenGORMRepo) GetAllForUser(ctx context.Context, userId entities.Id, filter entities.ApiTokenFilter) ([]entities.ApiToken, error) {
	gorm_tokens := make([]ApiToken, 0)

	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Where("owner_id = ?", userId).Preload("Owner").Find(&gorm_tokens).Error; err != nil {
		return nil, wrapGormError(err)
	}

	tokens := make([]entities.ApiToken, 0, len(gorm_tokens))
	for _, token := range gorm_tokens {
		tokens = append(tokens, apiTokenToEntity(token))
	}

	return tokens, nil
}

func (t *TokenGORMRepo) Update(ctx context.Context, token entities.ApiToken) (entities.ApiToken, error) {
	gorm_token := apiTokenFromEntity(token)
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Select("*").Updates(&gorm_token).Error; err != nil {
		return entities.ApiToken{}, wrapGormError(err)
	}

	return apiTokenToEntity(gorm_token), nil
}
