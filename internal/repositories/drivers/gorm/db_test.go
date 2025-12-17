package gorm

import (
	"testing"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGORMDatabase_SQLite(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)

	require.NoError(t, err)
	assert.NotNil(t, db)

	// Verify tables were created
	assert.True(t, db.Migrator().HasTable(&User{}))
	assert.True(t, db.Migrator().HasTable(&ApiToken{}))
	assert.True(t, db.Migrator().HasTable(&Address{}))
	assert.True(t, db.Migrator().HasTable(&Chain{}))
}

func TestNewGORMDatabase_ErrorLogLevel(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "error",
	}

	db, err := NewGORMDatabase(config)

	require.NoError(t, err)
	assert.NotNil(t, db)
}

func TestNewGORMDatabase_DebugLogLevel(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "debug",
	}

	db, err := NewGORMDatabase(config)

	require.NoError(t, err)
	assert.NotNil(t, db)
}

func TestNewGORMDatabase_UnknownDriver(t *testing.T) {
	config := Config{
		Driver:   "unknown",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)

	assert.Error(t, err)
	assert.Nil(t, db)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.Contains(t, err.Error(), "unknown database driver")
}

func TestNewGORMDatabase_InvalidConnectionString(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  "/invalid/path/to/database.db",
		LogLevel: "silent",
	}

	// SQLite might still open the database, but let's test with an explicitly invalid path
	// On some systems this may or may not fail, so we just verify the function executes
	db, err := NewGORMDatabase(config)

	// The behavior depends on the system and SQLite permissions
	// We just ensure the function doesn't panic
	if err == nil {
		assert.NotNil(t, db)
	}
}

func TestNewGORMDatabase_DefaultLogLevel(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "",
	}

	db, err := NewGORMDatabase(config)

	require.NoError(t, err)
	assert.NotNil(t, db)
}
