package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/controllers/rest"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

const (
	defaultConfigName = "config.json"
)

var (
	appVersion string
)

func main() {
	// parsing flags
	cfgName := flag.String("config", defaultConfigName, "path to the configuration file")
	version := flag.Bool("version", false, "")
	flag.Parse()

	if *version {
		fmt.Printf("Ovoo version: %s\n", appVersion)
		return
	}

	// load configuration
	cfg, err := config.NewParser(*cfgName)
	if err != nil {
		fmt.Printf("error parsing configuration: %s", err.Error())
		os.Exit(1)
	}

	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: getSLogLevel(cfg.String("log.level")),
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
	db_drv := cfg.String("api.database.type")
	db_config := cfg.StringMap("api.database.config")

	// initialize repo fabric
	repoFactory, err := factory.New(db_drv, db_config)
	if err != nil {
		slog.Error("error initializing repository", "err", err.Error())
		os.Exit(1)
	}

	// global context
	ctx := context.TODO()

	// initialize services
	domain := "alias-test.local"
	svcGw, err := makeServices(repoFactory, domain, dict)
	if err != nil {
		slog.Error("error initializing services gateway", "err", err.Error())
		os.Exit(1)
	}

	defaultAdminCfg := cfg.StringMap("api.default_admin")
	if len(defaultAdminCfg) > 0 {
		if err := makeDefaultAdmin(svcGw, defaultAdminCfg); err != nil {
			slog.Error("error creating default admin", "err", err.Error())
			os.Exit(1)
		}
	}

	// initialize REST controller
	listen_addr := cfg.String("api.listen_addr")
	if len(listen_addr) == 0 {
		listen_addr = rest.DefaultListenAddr
	}

	restApi, err := rest.New(listen_addr, logger, svcGw)
	if err != nil {
		slog.Error("error initializing rest api", "err", err.Error())
		os.Exit(1)
	}

	if err := restApi.Start(ctx); err != nil {
		slog.Error(err.Error())
	}
}
