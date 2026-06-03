package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Burmuley/ovoo/internal/applications/ovooclient"
	"github.com/Burmuley/ovoo/internal/applications/socketmap"
	"github.com/Burmuley/ovoo/internal/config"
)

func startSocketmap(cfg *config.SocketMapConfig) error {
	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: config.GetSLogLevel(cfg.Log.Level),
		},
	))
	slog.SetDefault(logger)

	network := cfg.Network
	if network == "" {
		network = socketmap.DefaultSocketmapNetwork
	}

	addr := cfg.ListenAddr
	if addr == "" {
		addr = socketmap.DefaultSocketmapAddr
	}

	cli, err := ovooclient.NewClient(cfg.Api.Addr, cfg.Api.AuthToken, cfg.Api.TLSSkipVerify, time.Duration(cfg.Api.Timeout))
	if err != nil {
		return err
	}

	app, err := socketmap.New(cfg.Network, cfg.ListenAddr, logger, cli)
	if err != nil {
		return err
	}

	return app.Start()
}
