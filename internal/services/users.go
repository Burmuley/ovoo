package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type UserCreateCmd struct {
	FirstName string
	LastName  string
	Login     string
	Password  *string
	Type      entities.UserType
}

type UserUpdateCmd struct {
	UserID    entities.Id
	FirstName *string
	LastName  *string
	Type      *entities.UserType
	Active    *bool
}

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
func (u *UsersService) Create(ctx context.Context, cuser entities.User, cmd UserCreateCmd) (entities.User, error) {
	if !canCreateUser(cuser) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	user := entities.User{
		Login:     cmd.Login,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Type:      cmd.Type,
		Active:    true,
	}

	user.ID = entities.NewId()
	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	{
		if cmd.Password != nil {
			var err error
			user.PasswordHash, err = entities.NewPasswordHash(user.PasswordHash)
			if err != nil {
				return entities.User{}, fmt.Errorf("%w: %w", entities.ErrGeneral, err)
			}
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

	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	err := u.repof.Users.Create(ctx, user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

// Update updates an existing user
func (u *UsersService) Update(ctx context.Context, cuser entities.User, cmd UserUpdateCmd) (entities.User, error) {
	user, err := u.repof.Users.GetById(ctx, cmd.UserID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
	}

	if !canUpdateUser(cuser, user.ID) {
		return entities.User{}, entities.ErrNotAuthorized
	}

	user.UpdatedBy = &cuser
	if cmd.FirstName != nil {
		user.FirstName = *cmd.FirstName
	}

	if cmd.LastName != nil {
		user.LastName = *cmd.LastName
	}

	if cmd.Type != nil {
		user.Type = *cmd.Type
	}

	if cmd.Active != nil {
		if canSetActiveUser(user, cuser) {
			if *cmd.Active == true && user.Active == false {
				user.Active = *cmd.Active
			} else if *cmd.Active == false && user.Active == true {
				if err := deactivatePrAddrsForUser(ctx, u.repof, user.ID); err != nil {
					return entities.User{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
				}

				if err := deactivateTokensForUser(ctx, u.repof, user.ID); err != nil {
					return entities.User{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
				}
				user.Active = *cmd.Active
			}
		} else {
			return entities.User{}, entities.ErrNotAuthorized
		}
	}

	if err := user.Validate(); err != nil {
		return entities.User{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := u.repof.Users.Update(ctx, user); err != nil {
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
func (u *UsersService) GetAll(ctx context.Context, cuser entities.User, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	filter.Count = true
	if cuser.Type != entities.AdminUser {
		var err error
		if filter, err = entities.NewUserFilter(map[string][]string{
			"page":      {strconv.Itoa(filter.Page)},
			"page_size": {strconv.Itoa(filter.PageSize)},
			"id":        {cuser.ID.String()},
		}); err != nil {
			return nil, entities.PaginationMetadata{}, err
		}
	}

	return u.repof.Users.GetAll(ctx, filter)
}
