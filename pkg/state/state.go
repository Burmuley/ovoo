package state

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrNilDialect     = errors.New("gorm dialect must be set")
	ErrRecordNotFound = errors.New("requested record was not found")
)

type State struct {
	dialect gorm.Dialector
	db      *gorm.DB
}

func NewState(params ...Parameter) (*State, error) {
	state := &State{}
	for _, param := range params {
		param(state)
	}

	if state.dialect == nil {
		return nil, ErrNilDialect
	}

	sdb, err := gorm.Open(state.dialect, &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})

	if err != nil {
		return nil, err
	}

	state.db = sdb

	// Initialize/Update database schema
	if err = state.db.AutoMigrate(&Alias{}); err != nil {
		return nil, err
	}
	if err = state.db.AutoMigrate(&ProtectedAddress{}); err != nil {
		return nil, err
	}
	if err = state.db.AutoMigrate(&ReplyAlias{}); err != nil {
		return nil, err
	}
	if err = state.db.AutoMigrate(&Sender{}); err != nil {
		return nil, err
	}

	return state, nil
}
