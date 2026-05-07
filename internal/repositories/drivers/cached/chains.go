package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type ChainsRepo struct {
	cache  cache.Cache
	config *config.ConfigCache
	repo   repositories.ChainReadWriter
}

func NewCachedChainsRepo(cache cache.Cache, repo repositories.ChainReadWriter, config *config.ConfigCache) (*ChainsRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &ChainsRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

func (c *ChainsRepo) GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	key := chainHashKey(hash)
	if chain, ok := getFromCache[entities.Chain](ctx, c.cache, key); ok {
		return chain, nil
	}
	chain, err := c.repo.GetByHash(ctx, hash)
	if err != nil {
		return entities.Chain{}, err
	}
	setInCache(ctx, c.cache, key, chain, durationSeconds(c.config.SingleItemTTL))
	return chain, nil
}

func (c *ChainsRepo) GetByFilters(ctx context.Context, filter entities.ChainFilter) ([]entities.Chain, error) {
	key := chainListKey(filter)
	if chains, ok := getFromCache[[]entities.Chain](ctx, c.cache, key); ok {
		return chains, nil
	}
	chains, err := c.repo.GetByFilters(ctx, filter)
	if err != nil {
		return nil, err
	}
	setInCache(ctx, c.cache, key, chains, durationSeconds(c.config.ListTTL))
	return chains, nil
}

func (c *ChainsRepo) Create(ctx context.Context, chain entities.Chain) error {
	if err := c.repo.Create(ctx, chain); err != nil {
		return err
	}
	evict(ctx, c.cache, chainHashKey(chain.Hash))
	evictPrefix(ctx, c.cache, chainListPrefix())
	return nil
}

func (c *ChainsRepo) BatchCreate(ctx context.Context, chains []entities.Chain) error {
	if err := c.repo.BatchCreate(ctx, chains); err != nil {
		return err
	}
	evictPrefix(ctx, c.cache, "chain:")
	return nil
}

func (c *ChainsRepo) Delete(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	chain, err := c.repo.Delete(ctx, hash)
	if err != nil {
		return entities.Chain{}, err
	}
	evict(ctx, c.cache, chainHashKey(hash))
	evictPrefix(ctx, c.cache, chainListPrefix())
	return chain, nil
}

func (c *ChainsRepo) BatchDelete(ctx context.Context, hashes []entities.Hash) error {
	if err := c.repo.BatchDelete(ctx, hashes); err != nil {
		return err
	}
	evictPrefix(ctx, c.cache, "chain:")
	return nil
}
