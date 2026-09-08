package main

import (
	"embed"
	"os"
	"strings"

	"github.com/Burmuley/ovoo/internal/config"
)

//go:embed data/templates/**
var templatesData embed.FS

func loadPrAddrNotifyTmpl(name string) (string, error) {
	if len(strings.TrimSpace(name)) == 0 {
		return loadPrAddrNotifyTmplEmbed(config.DefaultPrAddrNotifyTmplPath)
	}

	return loadPrAddrNotifyTmplFile(name)
}

func loadPrAddrNotifyTmplEmbed(name string) (string, error) {
	bs, err := templatesData.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func loadPrAddrNotifyTmplFile(name string) (string, error) {
	bs, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}
