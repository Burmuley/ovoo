package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/applications/milter"
	"github.com/Burmuley/ovoo/internal/config"
)

func startMilter(cfg config.MilterConfig) error {
	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: config.GetSLogLevel(cfg.Log.Level),
		},
	))
	slog.SetDefault(logger)

	// initialize Milter controller
	listen_addr := cfg.ListenAddr
	if len(listen_addr) == 0 {
		listen_addr = milter.DefaultListenAddr
	}

	ovooApiAddr := cfg.Api.Addr
	ovooApiToken := cfg.Api.AuthToken
	if ovooApiToken == "" {
		return errors.New("missing 'auth_token' configuration parameter")
	}
	ovooClient, err := milter.NewOvooClient(ovooApiAddr, ovooApiToken, cfg.Api.TlsSkipVerify, cfg.Domain)
	if err != nil {
		return fmt.Errorf("error creating Ovoo API client: %w", err)
	}
	ctrl, _ := milter.New(listen_addr, logger, ovooClient)
	return ctrl.Start()
}
