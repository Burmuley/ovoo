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
	gorm_user := UserFEntity(user)
	if err := u.db.WithContext(ctx).Model(&User{}).Create(&gorm_user).Error; err != nil {
		return wrapGormError(err)
	}

	return nil
}

// Update modifies an existing user in the database.
func (u *UserGORMRepo) Update(ctx context.Context, user entities.User) error {
	gorm_user := UserFEntity(user)
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

	return UserTEntity(user), nil
}

// GetByLogin retrieves a user from the database by login (email).
func (u *UserGORMRepo) GetByLogin(ctx context.Context, login entities.Email) (entities.User, error) {
	user := User{}
	if err := u.db.WithContext(ctx).Model(&User{}).Where("login = ?", login).First(&user).Error; err != nil {
		return entities.User{}, wrapGormError(err)
	}

	return UserTEntity(user), nil
}

// GetAll retrieves all users from the database.
func (u *UserGORMRepo) GetAll(ctx context.Context) ([]entities.User, error) {
	gorm_users := make([]User, 0)
	if err := u.db.WithContext(ctx).Model(&User{}).Find(&gorm_users).Error; err != nil {
		return nil, wrapGormError(err)
	}

	users := make([]entities.User, 0, len(gorm_users))
	for _, user := range gorm_users {
		users = append(users, UserTEntity(user))
	}

	return users, nil
}
