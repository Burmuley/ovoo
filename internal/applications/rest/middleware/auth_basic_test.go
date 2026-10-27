package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

// stubUsersRepo implements repositories.UsersReadWriter and only answers GetByLogin;
// validateBasicAuth never reaches the other methods.
type stubUsersRepo struct {
	user entities.User
	err  error
}

func (s *stubUsersRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	if s.err != nil {
		return entities.User{}, s.err
	}
	if login != s.user.Login {
		return entities.User{}, entities.ErrNotFound
	}
	return s.user, nil
}

func (s *stubUsersRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	return entities.User{}, entities.ErrNotFound
}

func (s *stubUsersRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	return nil, entities.PaginationMetadata{}, nil
}

func (s *stubUsersRepo) Create(ctx context.Context, user entities.User) error     { return nil }
func (s *stubUsersRepo) BatchCreate(ctx context.Context, u []entities.User) error { return nil }
func (s *stubUsersRepo) Update(ctx context.Context, user entities.User) error     { return nil }
func (s *stubUsersRepo) Delete(ctx context.Context, cuser entities.User, id entities.Id) error {
	return nil
}

// newTestGateway builds a gateway carrying only the users service. validateBasicAuth uses
// nothing else, and a struct literal avoids services.New's all-fields-set check.
func newTestGateway(t *testing.T, user entities.User) *services.ServiceGateway {
	t.Helper()
	usersSvc, err := services.NewUsersService(&factory.RepoFactory{Users: &stubUsersRepo{user: user}})
	require.NoError(t, err)
	return &services.ServiceGateway{Users: usersSvc}
}

func testAuthUser(t *testing.T, passwordHash string) entities.User {
	t.Helper()
	return entities.User{
		ID:           entities.NewId(),
		Type:         entities.RegularUser,
		Login:        "user@test.com",
		PasswordHash: passwordHash,
		Active:       true,
	}
}

func TestValidateBasicAuth_Success(t *testing.T) {
	hash, err := entities.NewPasswordHash("s3cr3t-p@ssw0rd")
	require.NoError(t, err)

	user := testAuthUser(t, hash)
	got, err := validateBasicAuth(context.Background(), user.Login, "s3cr3t-p@ssw0rd", newTestGateway(t, user))

	assert.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
}

func TestValidateBasicAuth_WrongPassword(t *testing.T) {
	hash, err := entities.NewPasswordHash("s3cr3t-p@ssw0rd")
	require.NoError(t, err)

	user := testAuthUser(t, hash)
	got, err := validateBasicAuth(context.Background(), user.Login, "wrong", newTestGateway(t, user))

	assert.Error(t, err)
	assert.Equal(t, entities.User{}, got)
}

func TestValidateBasicAuth_UnknownUser(t *testing.T) {
	hash, err := entities.NewPasswordHash("s3cr3t-p@ssw0rd")
	require.NoError(t, err)

	user := testAuthUser(t, hash)
	got, err := validateBasicAuth(context.Background(), "nobody@test.com", "s3cr3t-p@ssw0rd", newTestGateway(t, user))

	assert.Error(t, err)
	assert.Equal(t, entities.User{}, got)
}

// TestValidateBasicAuth_NoPasswordSet covers OIDC-only accounts: a user created without a
// password has an empty hash and must never be reachable through basic auth.
func TestValidateBasicAuth_NoPasswordSet(t *testing.T) {
	user := testAuthUser(t, "")

	for _, password := range []string{"", "anything"} {
		got, err := validateBasicAuth(context.Background(), user.Login, password, newTestGateway(t, user))

		assert.Error(t, err, "password %q was accepted for an OIDC-only account", password)
		assert.Contains(t, err.Error(), "password authentication is not enabled")
		assert.Equal(t, entities.User{}, got)
	}
}

// TestValidateBasicAuth_LegacyEmptyStringHash covers rows written by the release where
// UsersService.Create stored a bcrypt hash of the empty string: those accounts used to accept
// empty credentials, which was an authentication bypass.
func TestValidateBasicAuth_LegacyEmptyStringHash(t *testing.T) {
	legacy, err := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := testAuthUser(t, string(legacy))
	got, err := validateBasicAuth(context.Background(), user.Login, "", newTestGateway(t, user))

	assert.Error(t, err)
	assert.Equal(t, entities.User{}, got)
}

func TestValidateBasicAuth_InactiveUser(t *testing.T) {
	hash, err := entities.NewPasswordHash("s3cr3t-p@ssw0rd")
	require.NoError(t, err)

	user := testAuthUser(t, hash)
	user.Active = false

	got, err := validateBasicAuth(context.Background(), user.Login, "s3cr3t-p@ssw0rd", newTestGateway(t, user))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inactive user")
	assert.Equal(t, entities.User{}, got)
}
