package gorm

import (
	"context"
	"fmt"
	"strconv"

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

// GetAll retrieves all addresses from the database, with optional filters.
// Filters can be applied using the following keys:
// - "type": filter by address type (integer values)
// - "owner": filter by owner ID
// - "id": filter by address ID
// - "email": filter by email address
// - "service_name": filter by service_name metadata field
// Returns a slice of entities.Address and an error, if any.
func (a *AddressGORMRepo) GetAll(ctx context.Context, filters map[string][]string) ([]entities.Address, error) {
	gorm_addrs := make([]Address, 0)
	stmt := a.db.WithContext(ctx).Model(&Address{})

	for filter, vals := range filters {
		switch filter {
		case "type":
			atypes := make([]int, 0, len(vals))
			for _, val := range vals {
				atype, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("%w: unsupported address type '%s'", entities.ErrValidation, val)
				}
				atypes = append(atypes, atype)
			}
			stmt.Where("type IN ?", atypes)
		case "owner":
			stmt.Where("owner_id IN ?", vals)
		case "id":
			stmt.Where("id IN ?", vals)
		case "email":
			stmt.Where("email IN ?", vals)
		case "service_name":
			for _, val := range vals {
				stmt.Or(datatypes.JSONQuery("metadata").Likes(val, "service_name"))
			}
		default:
			return nil, fmt.Errorf("%w: unsupported filter '%s'", entities.ErrValidation, filter)
		}
	}

	if err := stmt.Preload("ForwardAddress").Preload("Owner").Find(&gorm_addrs).Error; err != nil {
		return nil, wrapGormError(err)
	}

	addrs := make([]entities.Address, 0, len(gorm_addrs))
	for _, addr := range gorm_addrs {
		addrs = append(addrs, addressToEntity(addr))
	}

	return addrs, nil
}
