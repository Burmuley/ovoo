package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/controllers/milter"
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
		fmt.Printf("Ovoo Milter version: %s\n", appVersion)
		return
	}

	// load configuration
	cfg, err := config.NewParser(*cfgName, "milter")
	if err != nil {
		fmt.Printf("error parsing configuration: %s", err.Error())
		os.Exit(1)
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

	ovooApiAddr := cfg.String("api_addr")
	ovooClient := milter.NewOvooClient(ovooApiAddr)
	ctrl, _ := milter.New(listen_addr, logger, ovooClient)
	ctrl.Start(context.Background())
}
