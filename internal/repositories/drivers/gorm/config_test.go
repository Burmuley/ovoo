package gorm

import (
	"testing"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ImportMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		wantErr  bool
		errType  error
		validate func(*testing.T, *Config)
	}{
		{
			name: "valid configuration",
			input: map[string]string{
				"driver":            "sqlite",
				"connection_string": ":memory:",
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "sqlite", c.Driver)
				assert.Equal(t, ":memory:", c.ConnStr)
			},
		},
		{
			name: "missing driver",
			input: map[string]string{
				"connection_string": ":memory:",
			},
			wantErr: true,
			errType: entities.ErrConfiguration,
		},
		{
			name: "missing connection_string",
			input: map[string]string{
				"driver": "sqlite",
			},
			wantErr: true,
			errType: entities.ErrConfiguration,
		},
		{
			name:    "empty map",
			input:   map[string]string{},
			wantErr: true,
			errType: entities.ErrConfiguration,
		},
		{
			name: "extra fields are ignored",
			input: map[string]string{
				"driver":            "sqlite",
				"connection_string": ":memory:",
				"extra_field":       "value",
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "sqlite", c.Driver)
				assert.Equal(t, ":memory:", c.ConnStr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			err := c.ImportMap(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, c)
				}
			}
		})
	}
}
