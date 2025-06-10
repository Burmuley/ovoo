package gorm

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGORMDatabase creates a new GORM database connection based on the provided configuration.
// It supports SQLite as the database driver and automatically migrates the necessary tables.
// Returns a pointer to gorm.DB and an error if any occurred during the process.
func NewGORMDatabase(config Config) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch config.Driver {
	case "sqlite":
		dialect = sqlite.Open(config.ConnStr)
	default:
		return nil, fmt.Errorf("%w: unknown database driver '%s'", entities.ErrConfiguration, config.Driver)
	}

	var logLevel logger.LogLevel

	// Set GORM logger log level based on configuration.
	// Supported levels: "error" (shows errors), "debug" (shows info), default is silent.
	switch config.LogLevel {
	case "error":
		logLevel = logger.Error
	case "debug":
		logLevel = logger.Info
	default:
		logLevel = logger.Silent
	}
	gdb, err := gorm.Open(dialect, &gorm.Config{
		Logger:         logger.Default.LogMode(logLevel),
		TranslateError: true,
	})

	if err != nil {
		return nil, err
	}

	tables := []any{&User{}, &ApiToken{}, &Address{}, &Chain{}}
	for _, t := range tables {
		if err := gdb.AutoMigrate(t); err != nil {
			return nil, err
		}
	}

	return gdb, nil
}
