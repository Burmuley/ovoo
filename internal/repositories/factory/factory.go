package factory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Burmuley/ovoo/internal/config"
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
func New(
	repo_type string,
	repo_config map[string]string,
	defAdminCfg *config.ApiDefaultAdminConfig,
	logger *slog.Logger,
) (*RepoFactory, error) {
	switch repo_type {
	case "gorm":
		repo, err := newGormRepoFactory(repo_config)
		if err != nil {
			return nil, err
		}

		if defAdminCfg != nil {
			if err := handleDefaultAdmin(logger, repo, defAdminCfg); err != nil {
				return nil, err
			}
		}

		return repo, nil
	}

	return nil, fmt.Errorf("%w: unknown repository type", entities.ErrConfiguration)
}

func handleDefaultAdmin(logger *slog.Logger, repo *RepoFactory, defAdminCfg *config.ApiDefaultAdminConfig) error {
	adminUser := entities.User{
		FirstName:    defAdminCfg.FirstName,
		LastName:     defAdminCfg.LastName,
		Login:        defAdminCfg.Login,
		ID:           entities.NewId(),
		Type:         entities.AdminUser,
		PasswordHash: defAdminCfg.Password,
	}

	if err := repo.Users.Create(context.Background(), adminUser); err != nil {
		if errors.Is(err, entities.ErrDuplicateEntry) {
			logger.Info("default admin user already present in the repository, not creating")
			return nil
		} else {
			return err
		}
	}

	logger.Info("created default admin user")
	return nil
}
