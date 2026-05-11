package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type TokensRepo struct {
	cache  cache.Cache
	config *config.ConfigCache
	repo   repositories.TokensReadWriter
}

func NewCachedTokensRepo(cache cache.Cache, repo repositories.TokensReadWriter, config *config.ConfigCache) (*TokensRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &TokensRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

func (t *TokensRepo) GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error) {
	key := tokenIdKey(tokenId)
	if token, ok := getFromCache[entities.ApiToken](ctx, t.cache, key); ok {
		return token, nil
	}
	token, err := t.repo.GetById(ctx, tokenId)
	if err != nil {
		return entities.ApiToken{}, err
	}
	setInCache(ctx, t.cache, key, token, durationSeconds(t.config.SingleItemTTL))
	return token, nil
}

func (t *TokensRepo) GetAll(ctx context.Context, filter entities.ApiTokenFilter) ([]entities.ApiToken, error) {
	key := tokenUserListKey(filter)
	if tokens, ok := getFromCache[[]entities.ApiToken](ctx, t.cache, key); ok {
		return tokens, nil
	}
	tokens, err := t.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	setInCache(ctx, t.cache, key, tokens, durationSeconds(t.config.ListTTL))
	return tokens, nil
}

func (t *TokensRepo) Create(ctx context.Context, token entities.ApiToken) error {
	if err := t.repo.Create(ctx, token); err != nil {
		return err
	}
	evictPrefix(ctx, t.cache, tokenUserPrefix(token.Owner.ID))
	return nil
}

func (t *TokensRepo) Update(ctx context.Context, token entities.ApiToken) (entities.ApiToken, error) {
	updated, err := t.repo.Update(ctx, token)
	if err != nil {
		return entities.ApiToken{}, err
	}
	evict(ctx, t.cache, tokenIdKey(token.ID))
	evictPrefix(ctx, t.cache, tokenUserPrefix(token.Owner.ID))
	return updated, nil
}

func (t *TokensRepo) BatchCreate(ctx context.Context, tokens []entities.ApiToken) error {
	if err := t.repo.BatchCreate(ctx, tokens); err != nil {
		return err
	}
	seen := make(map[entities.Id]struct{})
	for _, tok := range tokens {
		if _, ok := seen[tok.Owner.ID]; !ok {
			evictPrefix(ctx, t.cache, tokenUserPrefix(tok.Owner.ID))
			seen[tok.Owner.ID] = struct{}{}
		}
	}
	return nil
}

func (t *TokensRepo) Delete(ctx context.Context, tokenId entities.Id) error {
	// Opportunistic cache lookup: if the token is already cached we can scope the
	// user-list eviction to that owner rather than clearing all user lists.
	cached, hasCached := getFromCache[entities.ApiToken](ctx, t.cache, tokenIdKey(tokenId))

	if err := t.repo.Delete(ctx, tokenId); err != nil {
		return err
	}

	evict(ctx, t.cache, tokenIdKey(tokenId))
	if hasCached {
		evictPrefix(ctx, t.cache, tokenUserPrefix(cached.Owner.ID))
	} else {
		evictPrefix(ctx, t.cache, "token:user:")
	}
	return nil
}

func (t *TokensRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	if err := t.repo.BatchDeleteById(ctx, ids); err != nil {
		return err
	}
	evictPrefix(ctx, t.cache, "token:")
	return nil
}

func (t *TokensRepo) BatchDeleteForUser(ctx context.Context, id entities.Id) error {
	if err := t.repo.BatchDeleteForUser(ctx, id); err != nil {
		return err
	}
	evictPrefix(ctx, t.cache, tokenUserPrefix(id))
	return nil
}
