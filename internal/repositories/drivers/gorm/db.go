package gorm

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new GORM database connection based on the provided configuration.
// It supports SQLite as the database driver and automatically migrates the necessary tables.
// Returns a pointer to gorm.DB and an error if any occurred during the process.
func NewDatabase(config config.ConfigDB) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch config.Config.GORM.Driver {
	case "sqlite":
		dialect = sqlite.Open(config.Config.GORM.ConnectionString)
	default:
		return nil, fmt.Errorf("%w: unknown GORM database driver '%s'", entities.ErrConfiguration, config.Config.GORM.Driver)
	}

	var logLevel logger.LogLevel

	// Set GORM logger log level based on configuration.
	// Supported levels: "error" (shows errors), "debug" (shows info), default is silent.
	switch config.LogLevel {
	case "error":
		logLevel = logger.Error
	case "debug":
		logLevel = logger.Info
	case "info":
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

	if err := gdb.AutoMigrate(&User{}, &ApiToken{}, &Address{}, &Chain{}); err != nil {
		return nil, err
	}

	return gdb, nil
}
