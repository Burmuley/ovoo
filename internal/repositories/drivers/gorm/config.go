package gorm

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
)

// Config represents the configuration for the GORM database connection.
type Config struct {
	Driver   string
	ConnStr  string
	LogLevel string
}

// ImportMap imports configuration from a map and sets the Driver and ConnStr fields.
// It returns an error if the required configuration keys are missing.
func (c *Config) ImportMap(m map[string]string) error {
	driver, ok := m["driver"]
	if !ok {
		return fmt.Errorf("%w: driver not set in the configuration", entities.ErrConfiguration)
	}

	connStr, ok := m["connection_string"]
	if !ok {
		return fmt.Errorf("%w: connection string not defined in the configuration", entities.ErrConfiguration)
	}

	c.Driver = driver
	c.ConnStr = connStr
	return nil
}
