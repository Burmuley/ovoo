package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/gorm"
)

// ChainsGORMRepo represents a GORM-based repository for managing Chain entities.
type ChainsGORMRepo struct {
	db *gorm.DB
}

// NewChainsGORMRepo creates a new instance of ChainsGORMRepo.
// It returns an error if the provided database connection is nil.
func NewChainsGORMRepo(db *gorm.DB) (repositories.ChainReadWriter, error) {
	if db == nil {
		return &ChainsGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}

	return &ChainsGORMRepo{db: db}, nil
}

// GetByHash retrieves a Chain entity by its hash.
// It returns the Chain entity and any error encountered during the process.
func (c *ChainsGORMRepo) GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	chain := Chain{}
	err := c.db.WithContext(ctx).Model(&Chain{}).Where("hash = ?", hash).Preload("ToAddress").Preload("FromAddress").Preload("OrigFromAddress").Preload("OrigToAddress").First(&chain).Error
	if err != nil {
		return entities.Chain{}, wrapGormError(err)
	}

	return chainToEntity(chain), nil
}

// Create adds a new Chain entity to the repository.
// It returns an error if the creation process fails.
func (c *ChainsGORMRepo) Create(ctx context.Context, chain entities.Chain) error {
	gorm_chain := chainFromEntity(chain)
	if err := c.db.WithContext(ctx).Model(&Chain{}).Create(&gorm_chain).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (c *ChainsGORMRepo) BatchCreate(ctx context.Context, chains []entities.Chain) error {
	gorm_chains := chainFromEntityList(chains)
	if err := c.db.WithContext(ctx).Model(&Chain{}).Create(&gorm_chains).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Delete removes a Chain entity from the repository based on its hash.
// It returns an error if the deletion process fails.
func (c *ChainsGORMRepo) Delete(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	chain, err := c.GetByHash(ctx, hash)
	if err != nil {
		return entities.Chain{}, wrapGormError(err)
	}
	if err := c.db.WithContext(ctx).Model(&Chain{}).Where("hash = ?", hash.String()).Unscoped().Delete(&Chain{}, hash.String()).Error; err != nil {
		return entities.Chain{}, wrapGormError(err)
	}

	return chain, nil
}
