package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
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
		log.Fatal("expected command with parameters")

	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("Ovoo version: %s\n", appVersion)
		return
	case "api":
		apiCmd.Parse(os.Args[2:])
		if err := startApi(*apiCfgName); err != nil {
			slog.Error(err.Error())
		}
	case "milter":
		milterCmd.Parse(os.Args[2:])
		if err := startMilter(*milterCfgName); err != nil {
			slog.Error(err.Error())
		}
	default:
		log.Fatal("supported commands: api, milter, version")
	}
}
