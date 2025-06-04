package gorm

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	"gorm.io/gorm"
)

// UserGORMRepo represents a GORM-based repository for user operations.
type UserGORMRepo struct {
	db *gorm.DB
}

// NewUserGORMRepo creates a new UserGORMRepo instance.
func NewUserGORMRepo(db *gorm.DB) (repositories.UsersReadWriter, error) {
	if db == nil {
		return &UserGORMRepo{}, fmt.Errorf("%w: database can not be nil", entities.ErrConfiguration)
	}

	return &UserGORMRepo{db: db}, nil
}

// Create adds a new user to the database.
func (u *UserGORMRepo) Create(ctx context.Context, user entities.User) error {
	gorm_user := userFromEntity(user)
	if err := u.db.WithContext(ctx).Model(&User{}).Create(&gorm_user).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

func (u *UserGORMRepo) BatchCreate(ctx context.Context, users []entities.User) error {
	gorm_users := userFromEntityList(users)
	if err := u.db.WithContext(ctx).Model(&User{}).Create(&gorm_users).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Update modifies an existing user in the database.
func (u *UserGORMRepo) Update(ctx context.Context, user entities.User) error {
	gorm_user := userFromEntity(user)
	if err := u.db.WithContext(ctx).Model(&User{}).Select("*").Updates(&gorm_user).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Delete removes a user from the database by ID.
func (u *UserGORMRepo) Delete(ctx context.Context, id entities.Id) error {
	if _, err := u.GetById(ctx, id); err != nil {
		return wrapGormError(err)
	}

	return wrapGormError(u.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Unscoped().
		Delete(&User{Model: Model{ID: id.String()}}).Error)
}

// GetById retrieves a user from the database by ID.
func (u *UserGORMRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	user := User{}
	if err := u.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).First(&user).Error; err != nil {
		return entities.User{}, wrapGormError(err)
	}

	return userToEntity(user), nil
}

// GetByLogin retrieves a user from the database by login (email).
func (u *UserGORMRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	user := User{}
	if err := u.db.WithContext(ctx).Model(&User{}).Where("login = ?", login).First(&user).Error; err != nil {
		return entities.User{}, wrapGormError(err)
	}

	return userToEntity(user), nil
}

// GetAll retrieves all users from the database.
func (u *UserGORMRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	gorm_users := make([]User, 0)
	stmt := u.db.WithContext(ctx).Model(&User{})
	count := applyUserFilter(stmt, filter)
	if err := stmt.Find(&gorm_users).Error; err != nil {
		return nil, entities.PaginationMetadata{}, wrapGormError(err)
	}

	users := make([]entities.User, 0, len(gorm_users))
	for _, user := range gorm_users {
		users = append(users, userToEntity(user))
	}

	return users, entities.GetPaginationMetadata(filter.Page, filter.PageSize, *count), nil
}

func applyUserFilter(stmt *gorm.DB, filter entities.UserFilter) *int64 {
	if filter.Ids != nil {
		stmt.Where("id IN ?", filter.Ids)
	}

	if filter.Types != nil {
		stmt.Where("type IN ?", filter.Types)
	}

	if filter.Logins != nil {
		stmt.Where("login IN ?", filter.Logins)
	}

	var count int64
	stmt.Count(&count)
	stmt.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	return &count
}
