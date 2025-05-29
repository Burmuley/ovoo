package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/applications/rest"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

func makeServices(repoFactory *factory.RepoFactory, domain string, dict []string) (*services.ServiceGateway, error) {
	aliases, err := services.NewAliasesService(domain, dict, repoFactory)
	if err != nil {
		return nil, fmt.Errorf("initializing aliases service: %w", err)
	}

	prAddrs, err := services.NewProtectedAddrService(repoFactory)
	if err != nil {
		return nil, fmt.Errorf("initializing protected addresses service: %w", err)
	}

	chains, err := services.NewChainsService(domain, repoFactory)
	if err != nil {
		return nil, fmt.Errorf("initializing chains service: %w", err)
	}

	users, err := services.NewUsersService(repoFactory)
	if err != nil {
		return nil, fmt.Errorf("initializing users service: %w", err)
	}

	tokens, err := services.NewApiTokensService(repoFactory)
	if err != nil {
		return nil, fmt.Errorf("initializing api tokens service: %w", err)
	}

	svcGw, err := services.New(aliases, prAddrs, chains, users, tokens)
	if err != nil {
		return nil, fmt.Errorf("initializing services gateway: %w", err)
	}

	return svcGw, nil
}

// func makeDefaultAdmin(svcGw *services.ServiceGateway, admin config.ApiDefaultAdminConfig) error {
// 	adminUser := entities.User{
// 		FirstName:    admin.FirstName,
// 		LastName:     admin.LastName,
// 		Login:        admin.Login,
// 		ID:           entities.NewId(),
// 		Type:         entities.AdminUser,
// 		PasswordHash: admin.Password,
// 	}
// 	if _, err := svcGw.Users.CreatePriv(context.Background(), adminUser); err != nil {
// 		if errors.Is(err, entities.ErrDuplicateEntry) {
// 			slog.Info("default admin user already present in the repository, not creating")
// 			return nil
// 		} else {
// 			return err
// 		}
// 	}

// 	slog.Info("created default admin user")
// 	return nil
// }

func startApi(cfg config.ApiConfig) error {
	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: config.GetSLogLevel(cfg.Log.Level),
		},
	))
	slog.SetDefault(logger)

	// load words dictionary
	dict, err := loadDict()
	if err != nil {
		slog.Error("error loading dictionary", "error", err)
		os.Exit(1)
	}

	// database configuration
	db_drv := cfg.Database.DBType
	db_config := map[string]string{
		"driver":            cfg.Database.Config.Driver,
		"connection_string": cfg.Database.Config.ConnectionString,
	}

	// initialize repo fabric
	repoFactory, err := factory.New(db_drv, db_config, &cfg.DefaultAdmin, logger)
	if err != nil {
		return fmt.Errorf("error initializing repository: %w", err)
	}

	// global context
	ctx := context.TODO()

	// initialize services
	svcGw, err := makeServices(repoFactory, cfg.Domain, dict)
	if err != nil {
		return fmt.Errorf("error initializing services gateway: %w", err)
	}

	// if len(cfg.DefaultAdmin.Login) > 0 {
	// 	if err := makeDefaultAdmin(svcGw, cfg.DefaultAdmin); err != nil {
	// 		return fmt.Errorf("error creating default admin: %w", err)
	// 	}
	// }

	// initialize REST controller
	listen_addr := cfg.ListenAddr
	if len(listen_addr) == 0 {
		listen_addr = rest.DefaultListenAddr
	}

	restApi, err := rest.New(listen_addr, logger, svcGw, cfg.Tls.Key, cfg.Tls.Cert, cfg.OIDC)
	if err != nil {
		return fmt.Errorf("error initializing rest api: %w", err)
	}

	return restApi.Start(ctx)
}
