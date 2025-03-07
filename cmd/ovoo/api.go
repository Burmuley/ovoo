package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/controllers/rest"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

func makeServices(repoFactory *factory.RepoFactory, domain string, dict []string) (*services.ServiceGateway, error) {
	var err error
	svcGw := services.ServiceGateway{}

	if svcGw.Aliases, err = services.NewAliasesService(domain, dict, repoFactory); err != nil {
		return nil, fmt.Errorf("initializing aliases service: %w", err)
	}

	if svcGw.PrAddrs, err = services.NewProtectedAddrService(repoFactory); err != nil {
		return nil, fmt.Errorf("initializing protected addresses service: %w", err)
	}

	if svcGw.Chains, err = services.NewChainsService(domain, repoFactory); err != nil {
		return nil, fmt.Errorf("initializing chains service: %w", err)
	}

	if svcGw.Users, err = services.NewUsersService(repoFactory); err != nil {
		return nil, fmt.Errorf("initializing users service: %w", err)
	}

	return &svcGw, nil
}

func makeDefaultAdmin(svcGw *services.ServiceGateway, admin map[string]string) error {
	adminUser := entities.User{
		FirstName:    admin["firstName"],
		LastName:     admin["lastName"],
		Login:        admin["login"],
		ID:           entities.NewId(),
		Type:         entities.AdminUser,
		PasswordHash: admin["password"],
	}
	if _, err := svcGw.Users.Create(context.Background(), adminUser); err != nil {
		if errors.Is(err, entities.ErrDuplicateEntry) {
			slog.Info("default admin user already present in the repository, not creating")
			return nil
		} else {
			return err
		}
	}

	slog.Info("created default admin user")
	return nil
}

func startApi(cfgPath string) error {
	// load configuration
	cfg, err := config.NewParser(cfgPath, "api")
	if err != nil {
		return fmt.Errorf("error parsing configuration: %s", err.Error())
	}

	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: config.GetSLogLevel(cfg.String("log.level")),
		},
	))
	slog.SetDefault(logger)

	// load words dictionary
	dict, err := loadDict()
	if err != nil {
		slog.Error("error loading dictionary", "err", err.Error())
		os.Exit(1)
	}

	// database configuration
	db_drv := cfg.String("database.type")
	db_config := cfg.StringMap("database.config")

	// initialize repo fabric
	repoFactory, err := factory.New(db_drv, db_config)
	if err != nil {
		return fmt.Errorf("error initializing repository", "err", err.Error())
	}

	// global context
	ctx := context.TODO()

	// initialize services
	domain := "alias-test.local"
	svcGw, err := makeServices(repoFactory, domain, dict)
	if err != nil {
		return fmt.Errorf("error initializing services gateway", "err", err.Error())
	}

	defaultAdminCfg := cfg.StringMap("default_admin")
	if len(defaultAdminCfg) > 0 {
		if err := makeDefaultAdmin(svcGw, defaultAdminCfg); err != nil {
			return fmt.Errorf("error creating default admin", "err", err.Error())
		}
	}

	// initialize REST controller
	listen_addr := cfg.String("listen_addr")
	if len(listen_addr) == 0 {
		listen_addr = rest.DefaultListenAddr
	}

	restApi, err := rest.New(listen_addr, logger, svcGw, cfg.String("tls.key"), cfg.String("tls.cert"))
	if err != nil {
		return fmt.Errorf("error initializing rest api", "err", err.Error())
	}

	return restApi.Start(ctx)
}
