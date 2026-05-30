package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type CustomDomainRepo struct {
	cache  cache.Cache
	config *config.ConfigCache
	repo   repositories.CustomDomainsReadWriter
}

func NewCachedCustomDomainRepo(cache cache.Cache, repo repositories.CustomDomainsReadWriter, config *config.ConfigCache) (*CustomDomainRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &CustomDomainRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

type domainListResult struct {
	Domains []entities.CustomDomain     `json:"domains"`
	Meta    entities.PaginationMetadata `json:"meta"`
}

func (cd CustomDomainRepo) GetById(ctx context.Context, id entities.Id) (entities.CustomDomain, error) {
	key := customDomainIdKey(id)
	if domain, ok := getFromCache[entities.CustomDomain](ctx, cd.cache, key); ok {
		return domain, nil
	}
	domain, err := cd.repo.GetById(ctx, id)
	if err != nil {
		return entities.CustomDomain{}, err
	}
	setInCache(ctx, cd.cache, key, domain, durationSeconds(cd.config.SingleItemTTL))
	return domain, nil
}

func (cd CustomDomainRepo) GetAll(ctx context.Context, filter entities.CustomDomainFilter) ([]entities.CustomDomain, entities.PaginationMetadata, error) {
	key := customDomainKeyList(filter)
	if result, ok := getFromCache[domainListResult](ctx, cd.cache, key); ok {
		return result.Domains, result.Meta, nil
	}
	domains, meta, err := cd.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}
	setInCache(ctx, cd.cache, key, domainListResult{Domains: domains, Meta: meta}, durationSeconds(cd.config.ListTTL))
	return domains, meta, nil
}

func (cd CustomDomainRepo) Create(ctx context.Context, domain entities.CustomDomain) error {
	if err := cd.repo.Create(ctx, domain); err != nil {
		return err
	}
	evict(ctx, cd.cache, customDomainIdKey(domain.ID))
	evictPrefix(ctx, cd.cache, customDomainListPrefix())
	return nil
}

func (cd CustomDomainRepo) Update(ctx context.Context, domain entities.CustomDomain) (entities.CustomDomain, error) {
	var err error
	if domain, err = cd.repo.Update(ctx, domain); err != nil {
		return entities.CustomDomain{}, err
	}
	evict(ctx, cd.cache, customDomainIdKey(domain.ID))
	evictPrefix(ctx, cd.cache, customDomainListPrefix())
	return domain, nil
}

func (cd CustomDomainRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	// Opportunistic cache lookup: if we already have the domain cached we can
	// evict the domain key precisely without an extra DB round-trip.
	cached, hasCached := getFromCache[entities.CustomDomain](ctx, cd.cache, customDomainIdKey(id))

	if err := cd.repo.Delete(ctx, cuser, id); err != nil {
		return err
	}

	if hasCached {
		evict(ctx, cd.cache, customDomainIdKey(cached.ID))
	} else {
		evictPrefix(ctx, cd.cache, customDomainListPrefix())
	}

	return nil
}
