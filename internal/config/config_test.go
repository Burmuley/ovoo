package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

// LoadConfig[APIConfig] tests

func TestLoadConfig_APIConfig_Full(t *testing.T) {
	path := writeTempConfig(t, `{
		"api": {
			"listen_addr": "0.0.0.0:8808",
			"logging": {"destination": "stdout", "level": "debug"},
			"tls": {"cert": "/tls/cert.pem", "key": "/tls/key.pem"},
			"database": {
				"driver": "gorm",
				"log_level": "silent",
				"config": {"gorm": {"driver": "sqlite", "connection_string": ":memory:"}}
			},
			"oidc": {
				"google": {
					"client_id": "cid",
					"client_secret": "csecret",
					"issuer": "https://accounts.google.com"
				}
			},
			"default_admin": {
				"first_name": "Admin",
				"last_name": "User",
				"login": "admin",
				"password": "s3cr3t"
			},
			"cache": {
				"driver": "memory",
				"list_ttl": 60,
				"single_item_ttl": 300,
				"config": {}
			}
		}
	}`)

	cfg, err := LoadConfig[APIConfig](APISection, path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "0.0.0.0:8808", cfg.ListenAddr)
	assert.Equal(t, "stdout", cfg.Log.Destination)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "/tls/cert.pem", cfg.TLS.Cert)
	assert.Equal(t, "/tls/key.pem", cfg.TLS.Key)
	assert.Equal(t, "gorm", cfg.Database.Driver)
	assert.Equal(t, "silent", cfg.Database.LogLevel)
	assert.Equal(t, "sqlite", cfg.Database.Config.GORM.Driver)
	assert.Equal(t, ":memory:", cfg.Database.Config.GORM.ConnectionString)
	require.Contains(t, cfg.OIDC, "google")
	assert.Equal(t, "cid", cfg.OIDC["google"].ClientId)
	assert.Equal(t, "csecret", cfg.OIDC["google"].ClientSecret)
	assert.Equal(t, "https://accounts.google.com", cfg.OIDC["google"].Issuer)
	require.NotNil(t, cfg.DefaultAdmin)
	assert.Equal(t, "Admin", cfg.DefaultAdmin.FirstName)
	assert.Equal(t, "User", cfg.DefaultAdmin.LastName)
	assert.Equal(t, "admin", cfg.DefaultAdmin.Login)
	assert.Equal(t, "s3cr3t", cfg.DefaultAdmin.Password)
	require.NotNil(t, cfg.Cache)
	assert.Equal(t, "memory", cfg.Cache.CacheDriver)
	assert.Equal(t, 60, cfg.Cache.ListTTL)
	assert.Equal(t, 300, cfg.Cache.SingleItemTTL)
}

func TestLoadConfig_APIConfig_OptionalFieldsAbsent(t *testing.T) {
	path := writeTempConfig(t, `{
		"api": {
			"domains": ["example.com"],
			"database": {
				"driver": "gorm",
				"log_level": "silent",
				"config": {"gorm": {"driver": "sqlite", "connection_string": ":memory:"}}
			}
		}
	}`)

	cfg, err := LoadConfig[APIConfig](APISection, path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Nil(t, cfg.Cache)
	assert.Nil(t, cfg.DefaultAdmin)
	assert.Empty(t, cfg.OIDC)
	assert.Empty(t, cfg.ListenAddr)
}

func TestLoadConfig_APIConfig_EmptyJSON(t *testing.T) {
	path := writeTempConfig(t, `{}`)

	cfg, err := LoadConfig[APIConfig](APISection, path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Nil(t, cfg.Cache)
}

func TestLoadConfig_APIConfig_RedisCache(t *testing.T) {
	addr := "localhost:6379"
	path := writeTempConfig(t, `{
		"api": {
			"domain": "example.com",
			"database": {
				"driver": "gorm",
				"log_level": "silent",
				"config": {"gorm": {"driver": "sqlite", "connection_string": ":memory:"}}
			},
			"cache": {
				"driver": "redis",
				"list_ttl": 120,
				"single_item_ttl": 600,
				"config": {
					"redis": {
						"address": "`+addr+`",
						"db": 1,
						"protocol": 3,
						"username": "redisuser",
						"password": "redispass"
					}
				}
			}
		}
	}`)

	cfg, err := LoadConfig[APIConfig](APISection, path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Cache)
	assert.Equal(t, "redis", cfg.Cache.CacheDriver)
	require.NotNil(t, cfg.Cache.Config.Redis)
	require.NotNil(t, cfg.Cache.Config.Redis.Addr)
	assert.Equal(t, addr, *cfg.Cache.Config.Redis.Addr)
	assert.Equal(t, 1, cfg.Cache.Config.Redis.DB)
	assert.Equal(t, 3, cfg.Cache.Config.Redis.Protocol)
	require.NotNil(t, cfg.Cache.Config.Redis.Username)
	assert.Equal(t, "redisuser", *cfg.Cache.Config.Redis.Username)
	require.NotNil(t, cfg.Cache.Config.Redis.Password)
	assert.Equal(t, "redispass", *cfg.Cache.Config.Redis.Password)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig[APIConfig](APISection, "/nonexistent/path/config.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	path := writeTempConfig(t, `{not valid json}`)

	cfg, err := LoadConfig[APIConfig](APISection, path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

// LoadConfig[MilterConfig] tests

func TestLoadConfig_MilterConfig_Full(t *testing.T) {
	// LoadConfig always unmarshals from the "api" key regardless of type T,
	// so MilterConfig fields must be nested under "api" in the JSON.
	path := writeTempConfig(t, `{
		"milter": {
			"domains": ["example.com"],
			"listen_addr": "127.0.0.1:6785",
			"api": {
				"addr": "https://api.example.com",
				"auth_token": "secret-token",
				"tls_skip_verify": true
			},
			"log": {"destination": "stdout", "level": "info"}
		}
	}`)

	cfg, err := LoadConfig[MilterConfig](MilterSection, path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "127.0.0.1:6785", cfg.ListenAddr)
	assert.Equal(t, "https://api.example.com", cfg.Api.Addr)
	assert.Equal(t, "secret-token", cfg.Api.AuthToken)
	assert.True(t, cfg.Api.TLSSkipVerify)
	assert.Equal(t, "stdout", cfg.Log.Destination)
	assert.Equal(t, "info", cfg.Log.Level)
}

func TestLoadConfig_MilterConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig[MilterConfig](MilterSection, "/nonexistent/milter.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestLoadConfig_MilterConfig_InvalidJSON(t *testing.T) {
	path := writeTempConfig(t, `[invalid`)

	cfg, err := LoadConfig[MilterConfig](MilterSection, path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

// GetSLogLevel tests

func TestGetSLogLevel_KnownLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, GetSLogLevel(tc.input))
		})
	}
}

func TestGetSLogLevel_UnknownLevels(t *testing.T) {
	for _, input := range []string{"", "trace", "warn", "DEBUG", "INFO", "critical"} {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, slog.LevelInfo, GetSLogLevel(input))
		})
	}
}
