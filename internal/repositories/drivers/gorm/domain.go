package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChainsGORMRepo represents a GORM-based repository for managing Chain entities.
type CustomDomainGORMRepo struct {
	db *gorm.DB
}

// NewCustomDomainGORMRepo creates a new instance of CustomDomainGORMRepo.
// It returns an error if the provided database connection is nil.
func NewCustomDomainGORMRepo(db *gorm.DB) (repositories.CustomDomainsReadWriter, error) {
	if db == nil {
		return &CustomDomainGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}

	return &CustomDomainGORMRepo{db: db}, nil
}

func (cd CustomDomainGORMRepo) Create(ctx context.Context, domain entities.CustomDomain) error {
	gorm_domain := customDomainFromEntity(domain)
	if err := cd.db.WithContext(ctx).Model(&CustomDomain{}).Create(&gorm_domain).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (cd CustomDomainGORMRepo) Update(ctx context.Context, domain entities.CustomDomain) (entities.CustomDomain, error) {
	gorm_domain := customDomainFromEntity(domain)
	if err := cd.db.WithContext(ctx).Model(&CustomDomain{}).Select("*").Where("id = ?", domain.ID).Updates(&gorm_domain).Error; err != nil {
		return entities.CustomDomain{}, wrapGormError(err)
	}

	return customDomainToEntity(gorm_domain), nil
}

func (cd CustomDomainGORMRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	if _, err := cd.GetById(ctx, id); err != nil {
		return err
	}

	if err := cd.db.WithContext(ctx).Model(&CustomDomain{}).Where("id = ?", id).
		Updates(map[string]any{"updated_by_id": cuser.ID.String()}).Error; err != nil {
		return wrapGormError(err)
	}

	if err := cd.db.WithContext(ctx).Model(&CustomDomain{}).Unscoped().
		Delete(&CustomDomain{}, "id = ?", id.String()).Error; err != nil {
		return wrapGormError(err)
	}
	return nil
}

func (cd CustomDomainGORMRepo) GetById(ctx context.Context, id entities.Id) (entities.CustomDomain, error) {
	domain := CustomDomain{}
	if err := cd.db.WithContext(ctx).Preload(clause.Associations).Model(&CustomDomain{}).Where("id = ?", id).First(&domain).Error; err != nil {
		return entities.CustomDomain{}, wrapGormError(err)
	}

	return customDomainToEntity(domain), nil
}

func (cd CustomDomainGORMRepo) GetAll(ctx context.Context, filter entities.CustomDomainFilter) ([]entities.CustomDomain, entities.PaginationMetadata, error) {
	gorm_domains := make([]CustomDomain, 0)
	stmt := cd.db.WithContext(ctx).Model(&CustomDomain{})
	count := applyCustomDomainFilter(stmt, filter, true)
	if err := stmt.Preload(clause.Associations).Find(&gorm_domains).Error; err != nil {
		return nil, entities.PaginationMetadata{}, wrapGormError(err)
	}

	domains := customDomainToEntityList(gorm_domains)

	return domains, entities.GetPaginationMetadata(filter.Page, filter.PageSize, *count), nil
}

func applyCustomDomainFilter(stmt *gorm.DB, filter entities.CustomDomainFilter, doCount bool) *int64 {
	if filter.Active != nil {
		stmt.Where("active = ?", *filter.Active)
	}

	if filter.Verified != nil {
		stmt.Where("verified = ?", *filter.Verified)
	}

	if len(filter.Owners) > 0 {
		if filter.IncludeGlobal {
			stmt.Where("owner_id IN ? OR global = ?", filter.Owners, true)
		} else {
			stmt.Where("owner_id IN ?", filter.Owners)
		}
	} else if filter.IncludeGlobal {
		stmt.Where("global = ?", true)
	}

	var count int64 = 0
	if doCount {
		stmt.Count(&count)
	}

	if filter.Page != 0 && filter.PageSize != 0 {
		stmt.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}

	return &count
}
