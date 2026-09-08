package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddressVerifyGORMRepo implements the repositories.AddressVerifyReadWriter interface using GORM.
type AddressVerifyGORMRepo struct {
	db *gorm.DB
}

// NewAddressVerifyGORMRepo creates a new AddressGORMRepo instance.
func NewAddressVerifyGORMRepo(db *gorm.DB) (repositories.AddressVerifyReadWriter, error) {
	if db == nil {
		return &AddressVerifyGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}
	return &AddressVerifyGORMRepo{db: db}, nil
}

func (v *AddressVerifyGORMRepo) GetById(ctx context.Context, id entities.Id) (entities.AddressVerifyToken, error) {
	token := AddressVerifyToken{}
	if err := v.db.WithContext(ctx).Preload(clause.Associations).Model(&AddressVerifyToken{}).Where("id = ?", id).First(&token).Error; err != nil {
		return entities.AddressVerifyToken{}, wrapGormError(err)
	}

	return addrVerifyTokenTEntity(token), nil
}

func (v *AddressVerifyGORMRepo) GetByAddrId(ctx context.Context, addrId entities.Id) ([]entities.AddressVerifyToken, error) {
	tokens := make([]AddressVerifyToken, 0)
	if err := v.db.WithContext(ctx).Preload(clause.Associations).Model(&AddressVerifyToken{}).Where("addr_id = ?", addrId.String()).Find(&tokens).Error; err != nil {
		return nil, wrapGormError(err)
	}

	return addrVerifyTokenTEntityList(tokens), nil
}

func (v *AddressVerifyGORMRepo) Create(ctx context.Context, token entities.AddressVerifyToken) error {
	gormToken := addrVerifyTokenFEntity(token)
	if err := v.db.WithContext(ctx).Model(&AddressVerifyToken{}).Create(&gormToken).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (v *AddressVerifyGORMRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	if _, err := v.GetById(ctx, id); err != nil {
		return err
	}

	if err := v.db.WithContext(ctx).Model(&AddressVerifyToken{}).Where("id = ?", id).
		Updates(map[string]any{"updated_by_id": cuser.ID.String()}).Error; err != nil {
		return wrapGormError(err)
	}

	if err := v.db.WithContext(ctx).Model(&AddressVerifyToken{}).Unscoped().
		Delete(&AddressVerifyToken{}, "id = ?", id.String()).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}
