package config

import (
	"encoding/json"
	"io"
	"os"
)

//go:generate go tool github.com/Burmuley/dysconfig -schema=config_schema.json -output=config.gen.go -package=config

func LoadConfig(path string) (OvooConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return OvooConfig{}, err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return OvooConfig{}, err
	}

	config := OvooConfig{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return OvooConfig{}, err
	}

	return config, nil
}
