package main

import (
	"log/slog"
	"os"
	"time"

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

	cli, err := socketmap.NewClient(cfg.Api.Addr, cfg.Api.AuthToken, cfg.Api.TLSSkipVerify, time.Duration(cfg.Api.Timeout))
	if err != nil {
		return err
	}

	// return socketmap.ListenAndServe("unix", "/tmp/tstsocket.sock", handler)
	app, err := socketmap.New(cfg.Network, cfg.ListenAddr, logger, cli)
	if err != nil {
		return err
	}

	return app.Start()
}

// func handler(ctx context.Context, lookup, key string) (result string, found bool, err error) {
// 	_ = ctx
// 	slog.Info(fmt.Sprintf("sockmap handler: lookup=%s key=%s", lookup, key))
// 	return "YES", true, nil

// }
