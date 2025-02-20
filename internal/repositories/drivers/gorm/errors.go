package gorm

import (
	"errors"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"gorm.io/gorm"
)

// wrapGormError wraps GORM-specific errors into more generic application errors.
// It maps GORM's ErrRecordNotFound to entities.ErrNotFound,
// ErrDuplicatedKey to entities.ErrDuplicateEntry,
// and all other errors to entities.ErrGeneral.
func wrapGormError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %w", entities.ErrNotFound, err)
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("%w: %w", entities.ErrDuplicateEntry, err)
	}

	return fmt.Errorf("%w: %w", entities.ErrGeneral, err)
}
