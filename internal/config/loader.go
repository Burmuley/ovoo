package config

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func LoadConfig[T APIConfig | MilterConfig](section CfgSectionName, path string) (*T, error) {
	loader := koanf.New("/")
	if err := loader.Load(file.Provider(path), json.Parser()); err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}

	var cfg T
	if err := loader.Unmarshal(string(section), &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}

	return &cfg, nil
}
