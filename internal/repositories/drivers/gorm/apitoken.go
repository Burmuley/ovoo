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
func (t *TokenGORMRepo) Create(ctx context.Context, token entities.ApiToken) error {
	gorm_token := apiTokenFromEntity(token)
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Create(&gorm_token).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (t *TokenGORMRepo) BatchCreate(ctx context.Context, tokens []entities.ApiToken) error {
	gorm_tokens := apiTokenFromEntityList(tokens)
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Create(&gorm_tokens).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Delete removes an API token from the database based on its ID.
func (t *TokenGORMRepo) Delete(ctx context.Context, token_id entities.Id) error {
	if _, err := t.GetById(ctx, token_id); err != nil {
		return wrapGormError(err)
	}

	return wrapGormError(t.db.WithContext(ctx).Model(&ApiToken{}).Where("id = ?", token_id).Unscoped().
		Delete(&ApiToken{Model: Model{ID: token_id.String()}}).Error)
}

// GetById retrieves an API token from the database based on its ID.
func (t *TokenGORMRepo) GetById(ctx context.Context, token_id entities.Id) (entities.ApiToken, error) {
	token := ApiToken{}
	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Preload("Owner").Where("id = ?", token_id).First(&token).Error; err != nil {
		return entities.ApiToken{}, wrapGormError(err)
	}

	return apiTokenToEntity(token), nil
}

// GetAllForUser retrieves all API tokens associated with a given user.
func (t *TokenGORMRepo) GetAllForUser(ctx context.Context, user entities.User) ([]entities.ApiToken, error) {
	gorm_tokens := make([]ApiToken, 0)

	if err := t.db.WithContext(ctx).Model(&ApiToken{}).Preload("Owner").Where("owner_id = ?", user.ID).Find(&gorm_tokens).Error; err != nil {
		return nil, wrapGormError(err)
	}

	tokens := make([]entities.ApiToken, 0, len(gorm_tokens))
	for _, token := range gorm_tokens {
		tokens = append(tokens, apiTokenToEntity(token))
	}

	return tokens, nil
}
