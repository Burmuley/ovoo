package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type AddrsRepo struct {
	cache  cache.Cache
	config *config.APICacheConfig
	repo   repositories.AddressReadWriter
}

func NewCachedAddrsRepo(cache cache.Cache, repo repositories.AddressReadWriter, config *config.APICacheConfig) (*AddrsRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &AddrsRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

type addrListResult struct {
	Addresses []entities.Address          `json:"addresses"`
	Meta      entities.PaginationMetadata `json:"meta"`
}

func (a *AddrsRepo) GetById(ctx context.Context, id entities.Id) (entities.Address, error) {
	key := addrIdKey(id)
	if addr, ok := getFromCache[entities.Address](ctx, a.cache, key); ok {
		return addr, nil
	}
	addr, err := a.repo.GetById(ctx, id)
	if err != nil {
		return entities.Address{}, err
	}
	setInCache(ctx, a.cache, key, addr, durationSeconds(a.config.SingleItemTTL))
	return addr, nil
}

func (a *AddrsRepo) GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error) {
	key := addrEmailKey(email)
	if addrs, ok := getFromCache[[]entities.Address](ctx, a.cache, key); ok {
		return addrs, nil
	}
	addrs, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	setInCache(ctx, a.cache, key, addrs, durationSeconds(a.config.SingleItemTTL))
	return addrs, nil
}

func (a *AddrsRepo) GetAll(ctx context.Context, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error) {
	key := addrListKey(filter)
	if result, ok := getFromCache[addrListResult](ctx, a.cache, key); ok {
		return result.Addresses, result.Meta, nil
	}
	addrs, meta, err := a.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}
	setInCache(ctx, a.cache, key, addrListResult{Addresses: addrs, Meta: meta}, durationSeconds(a.config.ListTTL))
	return addrs, meta, nil
}

func (a *AddrsRepo) Create(ctx context.Context, address entities.Address) error {
	if err := a.repo.Create(ctx, address); err != nil {
		return err
	}
	evict(ctx, a.cache, addrIdKey(address.ID), addrEmailKey(address.Email))
	evictPrefix(ctx, a.cache, addrListPrefix())
	return nil
}

func (a *AddrsRepo) BatchCreate(ctx context.Context, addresses []entities.Address) error {
	if err := a.repo.BatchCreate(ctx, addresses); err != nil {
		return err
	}
	evictPrefix(ctx, a.cache, "addr:")
	return nil
}

func (a *AddrsRepo) Update(ctx context.Context, address entities.Address) error {
	if err := a.repo.Update(ctx, address); err != nil {
		return err
	}
	evict(ctx, a.cache, addrIdKey(address.ID), addrEmailKey(address.Email))
	evictPrefix(ctx, a.cache, addrListPrefix())
	return nil
}

func (a *AddrsRepo) DeleteById(ctx context.Context, id entities.Id) error {
	// Opportunistic cache lookup: if we already have the address cached we can
	// evict the email key precisely without an extra DB round-trip.
	cached, hasCached := getFromCache[entities.Address](ctx, a.cache, addrIdKey(id))

	if err := a.repo.DeleteById(ctx, id); err != nil {
		return err
	}

	evict(ctx, a.cache, addrIdKey(id))
	if hasCached {
		evict(ctx, a.cache, addrEmailKey(cached.Email))
	}
	evictPrefix(ctx, a.cache, addrListPrefix())
	return nil
}

func (a *AddrsRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	if err := a.repo.BatchDeleteById(ctx, ids); err != nil {
		return err
	}
	evictPrefix(ctx, a.cache, "addr:")
	return nil
}
