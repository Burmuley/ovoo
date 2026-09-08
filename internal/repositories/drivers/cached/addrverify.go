package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type AddrVerifyRepo struct {
	cache  cache.Cache
	config *config.ConfigCache
	repo   repositories.AddressVerifyReadWriter
}

func NewCachedAddrVerifyRepo(cache cache.Cache, repo repositories.AddressVerifyReadWriter, config *config.ConfigCache) (*AddrVerifyRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &AddrVerifyRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

func (v *AddrVerifyRepo) GetById(ctx context.Context, id entities.Id) (entities.AddressVerifyToken, error) {
	key := addrVerifyTokenIdKey(id)
	if token, ok := getFromCache[entities.AddressVerifyToken](ctx, v.cache, key); ok {
		return token, nil
	}
	token, err := v.repo.GetById(ctx, id)
	if err != nil {
		return entities.AddressVerifyToken{}, err
	}
	setInCache(ctx, v.cache, key, token, durationSeconds(v.config.SingleItemTTL))
	return token, nil
}

func (v *AddrVerifyRepo) GetByAddrId(ctx context.Context, addrId entities.Id) ([]entities.AddressVerifyToken, error) {
	key := addrVerifyTokenAddressKey(addrId)
	if tokens, ok := getFromCache[[]entities.AddressVerifyToken](ctx, v.cache, key); ok {
		return tokens, nil
	}
	tokens, err := v.repo.GetByAddrId(ctx, addrId)
	if err != nil {
		return nil, err
	}
	setInCache(ctx, v.cache, key, tokens, durationSeconds(v.config.SingleItemTTL))
	return tokens, nil
}

func (v *AddrVerifyRepo) Create(ctx context.Context, token entities.AddressVerifyToken) error {
	if err := v.repo.Create(ctx, token); err != nil {
		return err
	}
	evict(ctx, v.cache, addrVerifyTokenIdKey(token.ID), addrVerifyTokenAddressKey(token.AddrId))
	return nil
}

func (v *AddrVerifyRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	// Opportunistic cache lookup: if we already have the token cached we can
	// evict the verify token address keys precisely without an extra DB round-trip.
	cached, hasCached := getFromCache[entities.AddressVerifyToken](ctx, v.cache, addrVerifyTokenIdKey(id))

	if err := v.repo.Delete(ctx, cuser, id); err != nil {
		return err
	}

	evict(ctx, v.cache, addrIdKey(id))
	if hasCached {
		evict(ctx, v.cache, addrVerifyTokenAddressKey(cached.AddrId))
	} else {
		evictPrefix(ctx, v.cache, "addr_verify:")
	}
	return nil
}
