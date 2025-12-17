package gorm

import (
	"errors"
	"testing"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestWrapGormError(t *testing.T) {
	tests := []struct {
		name        string
		inputErr    error
		expectedErr error
	}{
		{
			name:        "wrap ErrRecordNotFound",
			inputErr:    gorm.ErrRecordNotFound,
			expectedErr: entities.ErrNotFound,
		},
		{
			name:        "wrap ErrDuplicatedKey",
			inputErr:    gorm.ErrDuplicatedKey,
			expectedErr: entities.ErrDuplicateEntry,
		},
		{
			name:        "wrap generic error",
			inputErr:    errors.New("some database error"),
			expectedErr: entities.ErrGeneral,
		},
		{
			name:        "wrap nil error",
			inputErr:    nil,
			expectedErr: entities.ErrGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := wrapGormError(tt.inputErr)

			assert.Error(t, wrappedErr)
			assert.ErrorIs(t, wrappedErr, tt.expectedErr)

			if tt.inputErr != nil {
				assert.ErrorIs(t, wrappedErr, tt.inputErr)
			}
		})
	}
}
