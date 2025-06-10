package services

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// UsersService represents the use case for user operations
type UsersService struct {
	repof *factory.RepoFactory
}

// NewUsersService creates a new UsersUsecase instance
func NewUsersService(repoFactory *factory.RepoFactory) (*UsersService, error) {
	if repoFactory == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}
	return &UsersService{repof: repoFactory}, nil
}

// Create creates a new user
func (u *UsersService) Create(ctx context.Context, cuser entities.User, user entities.User) (entities.User, error) {
	if !canCreateUser(cuser) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	user.ID = entities.NewId()
	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	{
		var err error
		user.PasswordHash, err = entities.NewPasswordHash(user.PasswordHash)
		if err != nil {
			return entities.User{}, fmt.Errorf("%w: %w", entities.ErrGeneral, err)
		}
	}

	user.UpdatedBy = &cuser
	err := u.repof.Users.Create(ctx, user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// Create creates a new user
// Should only be used in middleware or other system needs
// Should NEVER be user in external request handler
func (u *UsersService) CreatePriv(ctx context.Context, user entities.User) (entities.User, error) {
	user.ID = entities.NewId()
	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	{
		var err error
		user.PasswordHash, err = entities.NewPasswordHash(user.PasswordHash)
		if err != nil {
			return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}
	}

	err := u.repof.Users.Create(ctx, user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// Update updates an existing user
func (u *UsersService) Update(ctx context.Context, cuser entities.User, user entities.User) (entities.User, error) {
	if !canUpdateUser(cuser, user.ID) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	user.UpdatedBy = &cuser
	err := u.repof.Users.Update(ctx, user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// Delete removes a user by their ID
func (u *UsersService) Delete(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	user, err := u.repof.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, err
	}

	if !canDeleteUser(cuser, user) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	// delete protected addressed and related aliases/chains/reply_aliases
	if err := deletePrAddrsForUser(ctx, u.repof, user.ID); err != nil {
		return entities.User{}, err
	}

	// delete API tokens
	if err := u.repof.ApiTokens.BatchDeleteForUser(ctx, user.ID); err != nil {
		return entities.User{}, err
	}

	// delete user
	if err := u.repof.Users.Delete(ctx, user.ID); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// GetById retrieves a user by their ID
func (u *UsersService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if !canGetUser(cuser, id) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	user, err := u.repof.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

// GetByIdProv retrieves a user by their ID without checking access
// Should only be used in middleware or for other system needs
// Should NEVER be user in external request handler
func (u *UsersService) GetByIdPriv(ctx context.Context, cuser entities.User, id entities.Id) (entities.User, error) {
	if !canGetUser(cuser, id) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	if err := id.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	user, err := u.repof.Users.GetById(ctx, id)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

// GetByLogin retrieves a user by their login (email)
func (u *UsersService) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	user, err := u.repof.Users.GetByLogin(ctx, login)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// GetAll retrieves all users
func (u *UsersService) GetAll(ctx context.Context, cuser entities.User, filters map[string][]string) ([]entities.User, entities.PaginationMetadata, error) {
	var filter entities.UserFilter
	if cuser.Type == entities.AdminUser {
		{
			var err error
			if filter, err = entities.NewUserFilter(filters); err != nil {
				return nil, entities.PaginationMetadata{}, err
			}
		}
	} else {
		userFilters := map[string][]string{
			"page":      filters["page"],
			"page_size": filters["page_size"],
			"id":        {cuser.ID.String()},
		}
		var err error
		if filter, err = entities.NewUserFilter(userFilters); err != nil {
			return nil, entities.PaginationMetadata{}, err
		}
	}

	return u.repof.Users.GetAll(ctx, filter)
}
