package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/config"
)

const (
	defaultConfigName = "config.json"
)

var (
	appInfoVersion   string = "LOCAL"
	appInfoGitCommit string = "LOCAL"
	appInfoBuiltAt   string = "NOW"
)

func main() {
	// parsing flags
	apiCmd := flag.NewFlagSet("api", flag.ExitOnError)
	apiCfgName := apiCmd.String("config", defaultConfigName, "path to the configuration file")

	milterCmd := flag.NewFlagSet("milter", flag.ExitOnError)
	milterCfgName := milterCmd.String("config", defaultConfigName, "path to the configuration file")

	sockMapCmd := flag.NewFlagSet("socketmap", flag.ExitOnError)
	sockMapCfgName := sockMapCmd.String("config", defaultConfigName, "path to the configuration file")

	if len(os.Args) < 2 {
		printUsage(apiCmd, milterCmd, sockMapCmd)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("Version: %s\n", appInfoVersion)
		fmt.Printf("Git commit: %s\n", appInfoGitCommit)
		fmt.Printf("Build timestamp: %s\n", appInfoBuiltAt)
		return
	case "api":
		apiCmd.Parse(os.Args[2:])
		cfg, err := config.LoadConfig[config.APIConfig](config.APISection, *apiCfgName)
		if err != nil {
			log.Fatal(err)
		}

		cfg.Version = config.SystemVersion{
			Version:   appInfoVersion,
			GitCommit: appInfoGitCommit,
			BuiltAt:   appInfoBuiltAt,
		}
		if err := startApi(cfg); err != nil {
			slog.Error(err.Error())
		}

	case "milter":
		milterCmd.Parse(os.Args[2:])
		cfg, err := config.LoadConfig[config.MilterConfig](config.MilterSection, *milterCfgName)
		if err != nil {
			log.Fatal(err)
		}
		if err := startMilter(cfg); err != nil {
			slog.Error(err.Error())
		}
	case "socketmap":
		sockMapCmd.Parse(os.Args[2:])
		cfg, err := config.LoadConfig[config.SocketMapConfig](config.SocketMapSection, *sockMapCfgName)
		if err != nil {
			log.Fatal(err)
		}
		if err := startSocketmap(cfg); err != nil {
			slog.Error(err.Error())
		}
	default:
		printUsage(apiCmd, milterCmd)
	}
}

func printUsage(flags ...*flag.FlagSet) {
	fmt.Println("Supported commands: api, milter, version")
	for _, f := range flags {
		f.Usage()
	}
	os.Exit(1)
}
