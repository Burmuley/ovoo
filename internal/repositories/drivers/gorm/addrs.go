package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if err := a.db.WithContext(ctx).Model(&Address{}).Select("*").Where("id = ?", address.ID).Updates(&gorm_addr).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// DeleteById removes an address from the database by its ID.
func (a *AddressGORMRepo) DeleteById(ctx context.Context, id entities.Id) error {
	if _, err := a.GetById(ctx, id); err != nil {
		return err
	}

	if err := a.db.WithContext(ctx).Model(&Address{}).Unscoped().
		Delete(&Address{}, "id = ?", id.String()).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

/*
BatchDeleteById deletes multiple addresses from the database by their IDs.

Parameters:
- ctx: Context for the database operation.
- ids: Slice of entities.Id representing the IDs of the addresses to delete.

The function performs a hard delete (unscoped) of all Address records matching the given IDs.
It returns an error if the deletion fails; otherwise, it returns nil.
*/
func (a *AddressGORMRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	if err := a.db.WithContext(ctx).Model(&Address{}).Unscoped().Delete(&Address{}, "id IN ?", ids).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (a *AddressGORMRepo) BatchUpdate(ctx context.Context, filter entities.AddressFilter, values entities.AddressBulkUpdateFields) error {
	updates := &Address{}
	if values.Active != nil {
		updates.Active = *values.Active
	}
	if values.MetadataComment != nil {
		updates.Metadata.Comment = *values.MetadataComment
	}
	if values.MetadataServiceName != nil {
		updates.Metadata.ServiceName = *values.MetadataServiceName
	}

	stmt := a.db.WithContext(ctx).Model(&Address{})
	_ = applyAddressFilter(stmt, filter)

	if err := stmt.Updates(updates).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// GetById retrieves an address from the database by its ID.
func (a *AddressGORMRepo) GetById(ctx context.Context, id entities.Id) (entities.Address, error) {
	addr := Address{}
	if err := a.db.WithContext(ctx).Preload(clause.Associations).Preload("ForwardAddress."+clause.Associations).Model(&Address{}).Where("id = ?", id).First(&addr).Error; err != nil {
		return entities.Address{}, wrapGormError(err)
	}

	return addressToEntity(addr), nil
}

// GetByEmail retrieves an address from the database by its email.
// It returns the address as an entities.Address and an error, if any.
func (a *AddressGORMRepo) GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error) {
	addrs := make([]Address, 0)
	if err := a.db.WithContext(ctx).Preload(clause.Associations).Preload("ForwardAddress."+clause.Associations).Model(&Address{}).Where("email = ?", email).Find(&addrs).Error; err != nil {
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
	if err := stmt.Preload(clause.Associations).Preload("ForwardAddress." + clause.Associations).Find(&gorm_addrs).Error; err != nil {
		return nil, entities.PaginationMetadata{}, wrapGormError(err)
	}

	addrs := make([]entities.Address, 0, len(gorm_addrs))
	for _, addr := range gorm_addrs {
		addrs = append(addrs, addressToEntity(addr))
	}

	return addrs, entities.GetPaginationMetadata(filter.Page, filter.PageSize, *count), nil
}

// applyAddressFilter appends WHERE conditions derived from filter onto stmt and
// returns the total matching row count (via COUNT(*)) captured before pagination
// is applied.
//
// stmt is mutated in place through GORM's method-chaining API. The caller is
// responsible for executing the final query (e.g. Find, Updates) after this
// function returns.
//
// Supported filter fields:
//   - Ids, Emails, Types, Owners, ForwardAddressIds — IN-list predicates.
//   - ServiceNames — per-value OR LIKE against metadata.service_name (JSON).
//   - Active — equality predicate; skipped when nil.
//   - Search — wildcard OR-group across email, metadata.service_name, and
//     metadata.comment; isolated in a sub-session to preserve correct grouping.
//   - Page / PageSize — Limit+Offset pagination, applied only when both are > 0.
//
// The returned count reflects all matching rows before pagination; pass it to
// entities.GetPaginationMetadata to compute last-page information.
// Callers that do not need pagination metadata (e.g. BatchUpdate) may discard
// the return value.
func applyAddressFilter(stmt *gorm.DB, filter entities.AddressFilter) *int64 {
	if len(filter.Ids) > 0 {
		stmt.Where("id IN ?", filter.Ids)
	}

	if len(filter.Emails) > 0 {
		stmt.Where("email IN ?", filter.Emails)
	}

	if len(filter.Types) > 0 {
		stmt.Where("type IN ?", filter.Types)
	}

	if len(filter.Owners) > 0 {
		stmt.Where("owner_id IN ?", filter.Owners)
	}

	if len(filter.ServiceNames) > 0 {
		for _, val := range filter.ServiceNames {
			stmt.Or(datatypes.JSONQuery("metadata").Likes(val, "service_name"))
		}
	}

	if len(filter.ForwardAddressIds) > 0 {
		stmt.Where("forward_address_id IN ?", filter.ForwardAddressIds)
	}

	if filter.Active != nil {
		stmt.Where("active = ?", *filter.Active)
	}

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		group := stmt.Session(&gorm.Session{NewDB: true}).
			Where("email LIKE ?", pattern).
			Or(datatypes.JSONQuery("metadata").Likes(pattern, "service_name")).
			Or(datatypes.JSONQuery("metadata").Likes(pattern, "comment"))
		stmt.Where(group)
	}

	var count int64 = 0
	stmt.Count(&count)
	if filter.Page != 0 && filter.PageSize != 0 {
		stmt.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}

	return &count
}
