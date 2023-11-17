package state

import (
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

type Parameter func(*State)

func WithSQLite(dsn string) func(*State) {
	return func(state *State) {
		state.dialect = sqlite.Open(dsn)
	}
}

func WithPostgreSQL(dsn string) func(*State) {
	return func(state *State) {
		state.dialect = postgres.Open(dsn)
	}
}

func WithMySQL(dsn string) func(*State) {
	return func(state *State) {
		state.dialect = mysql.Open(dsn)
	}
}
