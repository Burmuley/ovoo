package factory

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
)

// capturingUsersRepo implements repositories.UsersReadWriter and records the users created
// through it; handleDefaultAdmin only ever calls Create.
type capturingUsersRepo struct {
	created []entities.User
	err     error
}

func (c *capturingUsersRepo) Create(ctx context.Context, user entities.User) error {
	if c.err != nil {
		return c.err
	}
	c.created = append(c.created, user)
	return nil
}

func (c *capturingUsersRepo) BatchCreate(ctx context.Context, users []entities.User) error {
	return nil
}
func (c *capturingUsersRepo) Update(ctx context.Context, user entities.User) error { return nil }
func (c *capturingUsersRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	return nil
}
func (c *capturingUsersRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	return entities.User{}, entities.ErrNotFound
}
func (c *capturingUsersRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	return entities.User{}, entities.ErrNotFound
}
func (c *capturingUsersRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	return nil, entities.PaginationMetadata{}, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestHandleDefaultAdmin_HashesPassword guards the bug where the configured password was
// written straight into pwd_hash: it was stored in plain text and, being a non-bcrypt string,
// could never authenticate.
func TestHandleDefaultAdmin_HashesPassword(t *testing.T) {
	usersRepo := &capturingUsersRepo{}
	cfg := &config.ConfigDefaultAdmin{
		FirstName: "Admin",
		LastName:  "User",
		Login:     "admin",
		Password:  "$upeRP@s$w0rd",
	}

	err := handleDefaultAdmin(discardLogger(), &RepoFactory{Users: usersRepo}, cfg)
	require.NoError(t, err)
	require.Len(t, usersRepo.created, 1)

	admin := usersRepo.created[0]
	assert.Equal(t, cfg.Login, admin.Login)
	assert.Equal(t, entities.AdminUser, admin.Type)
	assert.True(t, admin.Active)
	assert.NotEmpty(t, admin.ID)

	assert.NotEqual(t, cfg.Password, admin.PasswordHash, "password stored in plain text")
	assert.True(t, entities.ValidPassword(cfg.Password, admin.PasswordHash),
		"the configured password does not validate against the stored hash")
	assert.False(t, entities.ValidPassword("", admin.PasswordHash),
		"empty password validates against the stored hash")
}

func TestHandleDefaultAdmin_EmptyPasswordSkipsCreation(t *testing.T) {
	for _, password := range []string{"", "   "} {
		usersRepo := &capturingUsersRepo{}
		cfg := &config.ConfigDefaultAdmin{
			FirstName: "Admin",
			LastName:  "User",
			Login:     "admin",
			Password:  password,
		}

		err := handleDefaultAdmin(discardLogger(), &RepoFactory{Users: usersRepo}, cfg)

		// an admin without a password could only sign in through OIDC, so none is created
		assert.NoError(t, err)
		assert.Empty(t, usersRepo.created, "an admin was created for password %q", password)
	}
}

// TestHandleDefaultAdmin_AlreadyExists covers restarts: the admin row is already there and a
// duplicate must not fail startup.
func TestHandleDefaultAdmin_AlreadyExists(t *testing.T) {
	usersRepo := &capturingUsersRepo{err: entities.ErrDuplicateEntry}
	cfg := &config.ConfigDefaultAdmin{
		FirstName: "Admin",
		LastName:  "User",
		Login:     "admin",
		Password:  "$upeRP@s$w0rd",
	}

	err := handleDefaultAdmin(discardLogger(), &RepoFactory{Users: usersRepo}, cfg)

	assert.NoError(t, err)
	assert.Empty(t, usersRepo.created)
}
