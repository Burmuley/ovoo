package factory

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/drivers/gorm"
)

// newGormRepoFactory creates a new RepoFactory instance using GORM as the database driver.
// It takes a configuration map and returns a pointer to RepoFactory and an error.
// The function initializes the database connection and sets up repositories for Users, ApiTokens, Address, and Chain.
func newGormRepoFactory(config config.ConfigDB) (*RepoFactory, error) {
	db, err := gorm.NewDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}

	repoFactory := &RepoFactory{}
	if repoFactory.Users, err = gorm.NewUserGORMRepo(db); err != nil {
		return nil, err
	}

	if repoFactory.ApiTokens, err = gorm.NewApiTokenGORMRepo(db); err != nil {
		return nil, err
	}

	if repoFactory.Address, err = gorm.NewAddressGORMRepo(db); err != nil {
		return nil, err
	}

	if repoFactory.Chain, err = gorm.NewChainsGORMRepo(db); err != nil {
		return nil, err
	}

	if repoFactory.Domain, err = gorm.NewCustomDomainGORMRepo(db); err != nil {
		return nil, err
	}

	return repoFactory, nil
}
