package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/applications/milter"
	"github.com/Burmuley/ovoo/internal/config"
)

func startMilter(cfgPath string) error {
	// load configuration
	cfg, err := config.NewParser(cfgPath, "milter")
	if err != nil {
		return fmt.Errorf("error parsing configuration: %w", err)
	}

	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: config.GetSLogLevel(cfg.String("log.level")),
		},
	))
	slog.SetDefault(logger)

	// initialize Milter controller
	listen_addr := cfg.String("listen_addr")
	if len(listen_addr) == 0 {
		listen_addr = milter.DefaultListenAddr
	}

	ovooApiAddr := cfg.String("api.addr")
	ovooApiToken := cfg.String("api.auth_token")
	if ovooApiToken == "" {
		return errors.New("missing 'auth_token' configuration parameter")
	}
	ovooClient := milter.NewOvooClient(ovooApiAddr, ovooApiToken, cfg.Bool("api.tls_skip_verify"))
	ctrl, _ := milter.New(listen_addr, logger, ovooClient)
	return ctrl.Start(context.Background())
}
