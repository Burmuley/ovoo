package cached

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

type UsersRepo struct {
	cache  cache.Cache
	config *config.ConfigCache
	repo   repositories.UsersReadWriter
}

func NewCachedUsersRepo(cache cache.Cache, repo repositories.UsersReadWriter, config *config.ConfigCache) (*UsersRepo, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache instance can not be empty", entities.ErrValidation)
	}

	if repo == nil {
		return nil, fmt.Errorf("%w: repository instance can not be empty", entities.ErrValidation)
	}

	if config == nil {
		return nil, fmt.Errorf("%w: cache config can not be empty", entities.ErrValidation)
	}

	return &UsersRepo{
		cache:  cache,
		repo:   repo,
		config: config,
	}, nil
}

type userListResult struct {
	Users []entities.User             `json:"users"`
	Meta  entities.PaginationMetadata `json:"meta"`
}

func (u *UsersRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	key := userListKey(filter)
	if result, ok := getFromCache[userListResult](ctx, u.cache, key); ok {
		return result.Users, result.Meta, nil
	}
	users, meta, err := u.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, entities.PaginationMetadata{}, err
	}
	setInCache(ctx, u.cache, key, userListResult{Users: users, Meta: meta}, durationSeconds(u.config.ListTTL))
	return users, meta, nil
}

func (u *UsersRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	key := userIdKey(id)
	if user, ok := getFromCache[entities.User](ctx, u.cache, key); ok {
		return user, nil
	}
	user, err := u.repo.GetById(ctx, id)
	if err != nil {
		return entities.User{}, err
	}
	setInCache(ctx, u.cache, key, user, durationSeconds(u.config.SingleItemTTL))
	return user, nil
}

func (u *UsersRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	key := userLoginKey(login)
	if user, ok := getFromCache[entities.User](ctx, u.cache, key); ok {
		return user, nil
	}
	user, err := u.repo.GetByLogin(ctx, login)
	if err != nil {
		return entities.User{}, err
	}
	setInCache(ctx, u.cache, key, user, durationSeconds(u.config.SingleItemTTL))
	return user, nil
}

func (u *UsersRepo) Create(ctx context.Context, user entities.User) error {
	if err := u.repo.Create(ctx, user); err != nil {
		return err
	}
	evictPrefix(ctx, u.cache, userListPrefix())
	return nil
}

func (u *UsersRepo) BatchCreate(ctx context.Context, users []entities.User) error {
	if err := u.repo.BatchCreate(ctx, users); err != nil {
		return err
	}
	evictPrefix(ctx, u.cache, userListPrefix())
	return nil
}

func (u *UsersRepo) Update(ctx context.Context, user entities.User) error {
	if err := u.repo.Update(ctx, user); err != nil {
		return err
	}
	evict(ctx, u.cache, userIdKey(user.ID), userLoginKey(user.Login))
	evictPrefix(ctx, u.cache, userListPrefix())
	return nil
}

func (u *UsersRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	// Opportunistic cache lookup: if the user is already cached we can evict the
	// login key precisely without an extra DB round-trip.
	cached, hasCached := getFromCache[entities.User](ctx, u.cache, userIdKey(id))

	if err := u.repo.Delete(ctx, cuser, id); err != nil {
		return err
	}

	evict(ctx, u.cache, userIdKey(id))
	if hasCached {
		evict(ctx, u.cache, userLoginKey(cached.Login))
	}
	evictPrefix(ctx, u.cache, userListPrefix())
	return nil
}
