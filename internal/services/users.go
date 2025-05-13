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
func (u *UsersService) Create(ctx context.Context, cuser entities.User, user entities.User) (entities.User, error) {
	if !canCreateUser(cuser) {
		return entities.User{}, fmt.Errorf("creating user: %w", entities.ErrNotAuthorized)
	}

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

// Create creates a new user
// Should only be used in middleware or for other system needs
// Should NEVER be user in external request handler
func (u *UsersService) CreatePriv(ctx context.Context, user entities.User) (entities.User, error) {
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
func (u *UsersService) Update(ctx context.Context, cuser entities.User, user entities.User) (entities.User, error) {
	if !canUpdateUser(cuser, user.ID) {
		return entities.User{}, fmt.Errorf("updating user: %w", entities.ErrNotAuthorized)
	}

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
func (u *UsersService) Delete(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if !canDeleteUser(cuser) {
		return entities.User{}, fmt.Errorf("deleting user: %w", entities.ErrNotAuthorized)
	}

	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("deleting user: %w", err)
	}

	user, err := u.repoFactory.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, fmt.Errorf("deleting user by id: %w", err)
	}

	if err := u.repoFactory.Users.Delete(ctx, user.ID); err != nil {
		return entities.User{}, fmt.Errorf("deleting user: %w", err)
	}

	return user, nil
}

// GetById retrieves a user by their ID
func (u *UsersService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if !canGetUser(cuser, id) {
		return entities.User{}, fmt.Errorf("gettin user: %w", entities.ErrNotAuthorized)
	}

	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("getting user by id: %w", err)
	}

	user, err := u.repoFactory.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}

// GetByIdProv retrieves a user by their ID without checking access
// Should only be used in middleware or for other system needs
// Should NEVER be user in external request handler
func (u *UsersService) GetByIdPriv(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if !canGetUser(cuser, id) {
		return entities.User{}, fmt.Errorf("gettin user priv: %w", entities.ErrNotAuthorized)
	}

	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("getting user priv by id: %w", err)
	}

	user, err := u.repoFactory.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, fmt.Errorf("getting user priv by id: %w", err)
	}
	return user, nil
}

// GetByLogin retrieves a user by their login (email)
func (u *UsersService) GetByLogin(ctx context.Context, login entities.Email) (entities.User, error) {
	if err := login.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("getting user by login: %w", err)
	}

	user, err := u.repoFactory.Users.GetByLogin(ctx, login)
	if err != nil {
		return entities.User{}, fmt.Errorf("getting user by login: %w", err)
	}

	return user, nil
}

// GetAll retrieves all users
func (u *UsersService) GetAll(ctx context.Context, cuser entities.User) ([]entities.User, error) {
	users, err := u.repoFactory.Users.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all users: %w", err)
	}

	return users, nil
}
