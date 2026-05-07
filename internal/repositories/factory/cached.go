package factory

import (
	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/repositories/drivers/cached"
)

func newCachedRepoFactory(cache cache.Cache, repoFactory *RepoFactory, config *config.ConfigCache) (*RepoFactory, error) {
	cachedRF := &RepoFactory{}

	{
		var err error
		if cachedRF.Address, err = cached.NewCachedAddrsRepo(cache, repoFactory.Address, config); err != nil {
			return nil, err
		}
	}

	{
		var err error
		if cachedRF.Chain, err = cached.NewCachedChainsRepo(cache, repoFactory.Chain, config); err != nil {
			return nil, err
		}
	}

	{
		var err error
		if cachedRF.ApiTokens, err = cached.NewCachedTokensRepo(cache, repoFactory.ApiTokens, config); err != nil {
			return nil, err
		}
	}

	{
		var err error
		if cachedRF.Users, err = cached.NewCachedUsersRepo(cache, repoFactory.Users, config); err != nil {
			return nil, err
		}
	}

	return cachedRF, nil
}
