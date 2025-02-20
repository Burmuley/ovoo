package factory

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/drivers/gorm"
)

// newGormRepoFactory creates a new RepoFactory instance using GORM as the database driver.
// It takes a configuration map and returns a pointer to RepoFactory and an error.
// The function initializes the database connection and sets up repositories for Users, ApiTokens, Address, and Chain.
func newGormRepoFactory(config map[string]string) (*RepoFactory, error) {
	dbCfg := &gorm.Config{}
	if err := dbCfg.ImportMap(config); err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}
	db, err := gorm.NewGORMDatabase(*dbCfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}

	fabric := &RepoFactory{}
	if fabric.Users, err = gorm.NewUserGORMRepo(db); err != nil {
		return nil, err
	}

	if fabric.ApiTokens, err = gorm.NewApiTokenGORMRepo(db); err != nil {
		return nil, err
	}

	if fabric.Address, err = gorm.NewAddressGORMRepo(db); err != nil {
		return nil, err
	}

	if fabric.Chain, err = gorm.NewChainsGORMRepo(db); err != nil {
		return nil, err
	}

	return fabric, nil
}
