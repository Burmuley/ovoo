package services

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// UsersService represents the use case for user operations
type UsersService struct {
	repoFactory *factory.RepoFactory
}

// NewUsersService creates a new UsersUsecase instance
func NewUsersService(repoFactory *factory.RepoFactory) (*UsersService, error) {
	if repoFactory == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}
	return &UsersService{repoFactory: repoFactory}, nil
}

// Create creates a new user
func (u *UsersService) Create(ctx context.Context, user entities.User) (entities.User, error) {
	user.ID = entities.NewId()
	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("creating user: %w", err)
	}

	{
		var err error
		user.PasswordHash, err = entities.NewPasswordHash(user.PasswordHash)
		if err != nil {
			return entities.User{}, fmt.Errorf("hashing user password: %w", err)
		}
	}

	err := u.repoFactory.Users.Create(ctx, user)
	if err != nil {
		return entities.User{}, fmt.Errorf("creating user: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (u *UsersService) Update(ctx context.Context, user entities.User) (entities.User, error) {
	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("updating user: %w", err)
	}

	err := u.repoFactory.Users.Update(ctx, user)
	if err != nil {
		return entities.User{}, fmt.Errorf("updating user: %w", err)
	}

	return user, nil
}

// Delete removes a user by their ID
func (u *UsersService) Delete(ctx context.Context, id entities.Id) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	err := u.repoFactory.Users.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	return nil
}

// GetById retrieves a user by their ID
func (u *UsersService) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("getting user by id: %w", err)
	}

	al, err := u.repoFactory.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, fmt.Errorf("getting user by id: %w", err)
	}
	return al, nil
}

// GetByLogin retrieves a user by their login (email)
func (u *UsersService) GetByLogin(ctx context.Context, login entities.Email) (entities.User, error) {
	if err := login.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("getting user by login: %w", err)
	}

	al, err := u.repoFactory.Users.GetByLogin(ctx, login)
	if err != nil {
		return entities.User{}, fmt.Errorf("getting user by login: %w", err)
	}
	return al, nil
}

// GetAll retrieves all users
func (u *UsersService) GetAll(ctx context.Context) ([]entities.User, error) {
	users, err := u.repoFactory.Users.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all users: %w", err)
	}

	return users, nil
}
