package factory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/cache/drivers/memory"
	"github.com/Burmuley/ovoo/internal/cache/drivers/redis"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
)

// RepoFactory represents a collection of repositories for different entities.
type RepoFactory struct {
	Users         repositories.UsersReadWriter
	Address       repositories.AddressReadWriter
	ApiTokens     repositories.TokensReadWriter
	Chain         repositories.ChainReadWriter
	Domain        repositories.CustomDomainsReadWriter
	AddressVerify repositories.AddressVerifyReadWriter
}

// New creates a new RepoFactory instance based on the provided repository type and configuration.
// It returns a pointer to RepoFabric and an error if the repository type is unknown.
func New(
	dbConfig config.ConfigDB,
	cacheConfig *config.ConfigCache,
	defAdminCfg *config.ConfigDefaultAdmin,
	logger *slog.Logger,
) (*RepoFactory, error) {
	var repoFactory *RepoFactory

	switch dbConfig.Driver {
	case "gorm":
		var err error
		repoFactory, err = newGormRepoFactory(dbConfig)
		if err != nil {
			return nil, err
		}

		if defAdminCfg != nil {
			if err := handleDefaultAdmin(logger, repoFactory, defAdminCfg); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("%w: unknown repository type", entities.ErrConfiguration)
	}

	if cacheConfig != nil {
		var cache cache.Cache
		var err error
		switch cacheConfig.CacheDriver {
		case "memory":
			cache, _ = memory.New()
		case "redis":
			cache, err = redis.New(*cacheConfig)
		default:
			return nil, fmt.Errorf("%w: unknown cache driver '%s'", entities.ErrConfiguration, cacheConfig.CacheDriver)
		}

		repoFactory, err = newCachedRepoFactory(cache, repoFactory, cacheConfig)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
		}
	}

	return repoFactory, nil

}

func handleDefaultAdmin(logger *slog.Logger, repo *RepoFactory, defAdminCfg *config.ConfigDefaultAdmin) error {
	// without a password the default admin could only ever sign in through OIDC, which makes
	// bootstrapping it pointless - skip it instead of creating an unusable account
	if strings.TrimSpace(defAdminCfg.Password) == "" {
		logger.Warn("default admin password is not configured, skipping default admin creation",
			"login", defAdminCfg.Login)
		return nil
	}

	pwdHash, err := entities.NewPasswordHash(defAdminCfg.Password)
	if err != nil {
		return fmt.Errorf("hashing default admin password: %w", err)
	}

	adminUser := entities.User{
		FirstName:    defAdminCfg.FirstName,
		LastName:     defAdminCfg.LastName,
		Login:        defAdminCfg.Login,
		ID:           entities.NewId(),
		Type:         entities.AdminUser,
		PasswordHash: pwdHash,
		Active:       true,
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
