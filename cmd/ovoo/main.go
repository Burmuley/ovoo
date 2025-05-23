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
	appVersion string
)

func main() {
	// parsing flags
	apiCmd := flag.NewFlagSet("api", flag.ExitOnError)
	apiCfgName := apiCmd.String("config", defaultConfigName, "path to the configuration file")

	milterCmd := flag.NewFlagSet("milter", flag.ExitOnError)
	milterCfgName := milterCmd.String("config", defaultConfigName, "path to the configuration file")

	if len(os.Args) < 2 {
		printUsage(apiCmd, milterCmd)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("Ovoo version: %s\n", appVersion)
		return
	case "api":
		apiCmd.Parse(os.Args[2:])
		config, err := config.LoadConfig(*apiCfgName)
		if err != nil {
			log.Fatal(err)
		}

		if err := startApi(config.Api); err != nil {
			slog.Error(err.Error())
		}
	case "milter":
		milterCmd.Parse(os.Args[2:])
		config, err := config.LoadConfig(*milterCfgName)
		if err != nil {
			log.Fatal(err)
		}
		if err := startMilter(config.Milter); err != nil {
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
