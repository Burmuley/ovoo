package factory

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

// RepoFactory represents a collection of repositories for different entities.
type RepoFactory struct {
	Users     repositories.UsersReadWriter
	Address   repositories.AddressReadWriter
	ApiTokens repositories.TokensReadWriter
	Chain     repositories.ChainReadWriter
}

// New creates a new RepoFactory instance based on the provided repository type and configuration.
// It returns a pointer to RepoFabric and an error if the repository type is unknown.
func New(repo_type string, repo_config map[string]string) (*RepoFactory, error) {
	switch repo_type {
	case "gorm":
		return newGormRepoFactory(repo_config)
	}

	return nil, fmt.Errorf("%w: unknown repository type", entities.ErrConfiguration)
}
