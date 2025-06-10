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

	if err := c.db.WithContext(ctx).Model(&Chain{}).
		Where("hash = ?", hash).
		Preload("ToAddress").
		Preload("FromAddress").
		Preload("OrigFromAddress").
		Preload("OrigToAddress").
		First(&chain).Error; err != nil {
		return entities.Chain{}, wrapGormError(err)
	}

	return chainToEntity(chain), nil
}

/*
GetByFilters retrieves a list of Chain entities based on the provided filter criteria.

Parameters:
  - ctx: The context used for controlling cancellation and timeouts.
  - filter: An instance of entities.ChainFilter defining the conditions and pagination for the query.

Returns:
  - A slice of entities.Chain that match the filter criteria.
  - An error if the query fails.

The function applies filtering and pagination options to the GORM query using applyChainFilter,
loads related address fields via Preload, and returns the mapped result entities.
*/
func (c *ChainsGORMRepo) GetByFilters(ctx context.Context, filter entities.ChainFilter) ([]entities.Chain, error) {
	chains := make([]Chain, 0)
	stmt := c.db.WithContext(ctx).Model(&Chain{})
	applyChainFilter(stmt, filter)
	if err := stmt.Preload("OrigFromAddress").
		Preload("OrigToAddress").
		Preload("FromAddress").
		Preload("ToAddress").
		Find(&chains).Error; err != nil {
		return nil, err
	}

	return chainToEntityList(chains), nil
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
	if err := c.db.WithContext(ctx).Unscoped().Delete(&Chain{}, "hash = ?", hash.String()).Error; err != nil {
		return entities.Chain{}, wrapGormError(err)
	}

	return chain, nil
}

func (c *ChainsGORMRepo) BatchDelete(ctx context.Context, hashes []entities.Hash) error {
	if err := c.db.WithContext(ctx).Unscoped().Delete(&Chain{}, "hash IN ?", hashes).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func applyChainFilter(stmt *gorm.DB, filter entities.ChainFilter) *int64 {
	if len(filter.OrigFromAddrIds) > 0 {
		stmt.Where("orig_from_address_id IN ?", filter.OrigFromAddrIds)
	}

	if len(filter.OrigToAddrIds) > 0 {
		stmt.Where("orig_to_address_id IN ?", filter.OrigToAddrIds)
	}

	if len(filter.FromAddrsIds) > 0 {
		stmt.Where("from_address_id IN ?", filter.FromAddrsIds)
	}

	if len(filter.ToAddrIds) > 0 {
		stmt.Where("to_address_id IN ?", filter.ToAddrIds)
	}

	var count int64 = 0
	stmt.Count(&count)
	if filter.Page != 0 && filter.PageSize != 0 {
		stmt.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}

	return &count
}
