package main

import (
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

func startApi(cfg *config.APIConfig) error {
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

	// initialize repo fabric
	repos, err := factory.New(cfg.Database, cfg.Cache, cfg.DefaultAdmin, logger)
	if err != nil {
		return fmt.Errorf("error initializing repository: %w", err)
	}

	// initialize services
	svcGw, err := makeServices(repos, cfg.Domain, dict)
	if err != nil {
		return fmt.Errorf("error initializing services gateway: %w", err)
	}

	// initialize REST controller
	listen_addr := cfg.ListenAddr
	if len(listen_addr) == 0 {
		listen_addr = rest.DefaultListenAddr
	}

	app, err := rest.New(listen_addr, logger, svcGw, cfg.TLS.Key, cfg.TLS.Cert, cfg.OIDC)
	if err != nil {
		return fmt.Errorf("error initializing rest api: %w", err)
	}

	return app.Start()
}
