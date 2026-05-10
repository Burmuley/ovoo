package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Burmuley/ovoo/internal/applications/milter"
	"github.com/Burmuley/ovoo/internal/config"
)

func startMilter(cfg *config.MilterConfig) error {
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

	apiAddr := cfg.Api.Addr
	apiToken := cfg.Api.AuthToken
	if apiToken == "" {
		return errors.New("missing 'auth_token' configuration parameter")
	}
	client, err := milter.NewClient(
		apiAddr, apiToken, cfg.Api.TLSSkipVerify, cfg.Domain,
		time.Duration(cfg.Api.Timeout)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("error creating Ovoo API client: %w", err)
	}
	app, _ := milter.New(listen_addr, logger, client)
	return app.Start()
}
