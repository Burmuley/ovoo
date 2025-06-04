package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AddressGORMRepo implements the repositories.AddressReadWriter interface using GORM.
type AddressGORMRepo struct {
	db *gorm.DB
}

// NewAddressGORMRepo creates a new AddressGORMRepo instance.
func NewAddressGORMRepo(db *gorm.DB) (repositories.AddressReadWriter, error) {
	if db == nil {
		return &AddressGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}
	return &AddressGORMRepo{db: db}, nil
}

// Create adds a new address to the database.
func (a *AddressGORMRepo) Create(ctx context.Context, address entities.Address) error {
	gorm_addr := addressFromEntity(address)
	if err := a.db.WithContext(ctx).Model(&Address{}).Create(&gorm_addr).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (a *AddressGORMRepo) BatchCreate(ctx context.Context, addresses []entities.Address) error {
	gorm_addrs := addressFromEntityList(addresses)
	if err := a.db.WithContext(ctx).Model(&Address{}).Create(&gorm_addrs).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Update modifies an existing address in the database.
func (a *AddressGORMRepo) Update(ctx context.Context, address entities.Address) error {
	gorm_addr := addressFromEntity(address)
	if err := a.db.WithContext(ctx).Model(&Address{}).Select("*").Updates(&gorm_addr).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// DeleteById removes an address from the database by its ID.
func (a *AddressGORMRepo) DeleteById(ctx context.Context, id entities.Id) error {
	if _, err := a.GetById(ctx, id); err != nil {
		return wrapGormError(err)
	}

	if err := a.db.WithContext(ctx).Model(&Address{}).Where("id = ?", id).Unscoped().
		Delete(&Address{Model: Model{ID: id.String()}}).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// GetById retrieves an address from the database by its ID.
func (a *AddressGORMRepo) GetById(ctx context.Context, id entities.Id) (entities.Address, error) {
	addr := Address{}
	if err := a.db.WithContext(ctx).Model(&Address{}).Where("id = ?", id).Preload("ForwardAddress").Preload("Owner").First(&addr).Error; err != nil {
		return entities.Address{}, wrapGormError(err)
	}

	return addressToEntity(addr), nil
}

// GetByEmail retrieves an address from the database by its email.
// It returns the address as an entities.Address and an error, if any.
func (a *AddressGORMRepo) GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error) {
	addrs := make([]Address, 0)
	if err := a.db.WithContext(ctx).Model(&Address{}).Where("email = ?", email).Preload("ForwardAddress").Preload("Owner").Find(&addrs).Error; err != nil {
		return []entities.Address{}, wrapGormError(err)
	}

	return addressToEntityList(addrs), nil
}

// GetAll retrieves all addresses from the database with pagination and filtering support.
// The entities.AddressFilter struct allows filtering by:
// - Ids: slice of address IDs to include
// - Emails: slice of email addresses to include
// - Types: slice of address types to include
// - Owners: slice of owner IDs to include
// - ServiceNames: slice of service names to match in metadata
// - Page: page number for pagination (1-based)
// - PageSize: number of items per page
// Returns a slice of entities.Address and an error, if any.
func (a *AddressGORMRepo) GetAll(ctx context.Context, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error) {
	gorm_addrs := make([]Address, 0)
	stmt := a.db.WithContext(ctx).Model(&Address{})
	count := applyAddressFilter(stmt, filter)
	if err := stmt.Preload("ForwardAddress").Preload("Owner").Find(&gorm_addrs).Error; err != nil {
		return nil, entities.PaginationMetadata{}, wrapGormError(err)
	}

	addrs := make([]entities.Address, 0, len(gorm_addrs))
	for _, addr := range gorm_addrs {
		addrs = append(addrs, addressToEntity(addr))
	}

	return addrs, entities.GetPaginationMetadata(filter.Page, filter.PageSize, *count), nil
}

func applyAddressFilter(stmt *gorm.DB, filter entities.AddressFilter) *int64 {
	if filter.Ids != nil {
		stmt.Where("id IN ?", filter.Ids)
	}

	if filter.Emails != nil {
		stmt.Where("email IN ?", filter.Emails)
	}

	if filter.Types != nil {
		stmt.Where("type IN ?", filter.Types)
	}

	if filter.Owners != nil {
		stmt.Where("owner_id IN ?", filter.Owners)
	}

	if filter.ServiceNames != nil {
		for _, val := range filter.ServiceNames {
			stmt.Or(datatypes.JSONQuery("metadata").Likes(val, "service_name"))
		}
	}
	var count int64
	stmt.Count(&count)
	stmt.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	return &count
}
