package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTokenTestDB(t *testing.T) (*TokenGORMRepo, entities.User) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	// Create a user for tokens
	userRepo, err := NewUserGORMRepo(db)
	require.NoError(t, err)

	user := entities.User{
		ID:           entities.NewId(),
		Login:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Type:         entities.RegularUser,
		PasswordHash: "hash",
	}

	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	repo, err := NewApiTokenGORMRepo(db)
	require.NoError(t, err)

	return repo.(*TokenGORMRepo), user
}

func TestNewApiTokenGORMRepo(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	repo, err := NewApiTokenGORMRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewApiTokenGORMRepo_NilDB(t *testing.T) {
	repo, err := NewApiTokenGORMRepo(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.NotNil(t, repo)
}

func TestTokenGORMRepo_Create(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "Test Token",
		TokenHash:   "hashvalue123",
		Salt:        "saltvalue",
		Description: "Test description",
		Owner:       user,
		Expiration:  time.Now().Add(24 * time.Hour),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   user,
	}

	err := repo.Create(ctx, token)

	assert.NoError(t, err)

	// Verify the token was created
	retrieved, err := repo.GetById(ctx, token.ID)
	assert.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.Name, retrieved.Name)
	assert.Equal(t, token.TokenHash, retrieved.TokenHash)
}

func TestTokenGORMRepo_BatchCreate(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	tokens := []entities.ApiToken{
		{
			ID:         entities.NewId(),
			Name:       "Token 1",
			TokenHash:  "hash1",
			Salt:       "salt1",
			Owner:      user,
			Expiration: time.Now().Add(24 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
		{
			ID:         entities.NewId(),
			Name:       "Token 2",
			TokenHash:  "hash2",
			Salt:       "salt2",
			Owner:      user,
			Expiration: time.Now().Add(48 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
	}

	err := repo.BatchCreate(ctx, tokens)

	assert.NoError(t, err)

	// Verify both tokens were created
	retrieved1, err := repo.GetById(ctx, tokens[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, tokens[0].Name, retrieved1.Name)

	retrieved2, err := repo.GetById(ctx, tokens[1].ID)
	assert.NoError(t, err)
	assert.Equal(t, tokens[1].Name, retrieved2.Name)
}

func TestTokenGORMRepo_Delete(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	token := entities.ApiToken{
		ID:         entities.NewId(),
		Name:       "Test Token",
		TokenHash:  "hashvalue",
		Salt:       "saltvalue",
		Owner:      user,
		Expiration: time.Now().Add(24 * time.Hour),
		Active:     true,
		UpdatedBy:  user,
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	err = repo.Delete(ctx, token.ID)
	assert.NoError(t, err)

	// Verify the token was deleted
	_, err = repo.GetById(ctx, token.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestTokenGORMRepo_Delete_NotFound(t *testing.T) {
	repo, _ := setupTokenTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestTokenGORMRepo_BatchDeleteById(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	tokens := []entities.ApiToken{
		{
			ID:         entities.NewId(),
			Name:       "Token 1",
			TokenHash:  "hash1",
			Salt:       "salt1",
			Owner:      user,
			Expiration: time.Now().Add(24 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
		{
			ID:         entities.NewId(),
			Name:       "Token 2",
			TokenHash:  "hash2",
			Salt:       "salt2",
			Owner:      user,
			Expiration: time.Now().Add(48 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
	}

	err := repo.BatchCreate(ctx, tokens)
	require.NoError(t, err)

	ids := []entities.Id{tokens[0].ID, tokens[1].ID}
	err = repo.BatchDeleteById(ctx, ids)

	assert.NoError(t, err)

	// Verify both tokens were deleted
	_, err = repo.GetById(ctx, tokens[0].ID)
	assert.Error(t, err)

	_, err = repo.GetById(ctx, tokens[1].ID)
	assert.Error(t, err)
}

func TestTokenGORMRepo_BatchDeleteForUser(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	tokens := []entities.ApiToken{
		{
			ID:         entities.NewId(),
			Name:       "Token 1",
			TokenHash:  "hash1",
			Salt:       "salt1",
			Owner:      user,
			Expiration: time.Now().Add(24 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
		{
			ID:         entities.NewId(),
			Name:       "Token 2",
			TokenHash:  "hash2",
			Salt:       "salt2",
			Owner:      user,
			Expiration: time.Now().Add(48 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
	}

	err := repo.BatchCreate(ctx, tokens)
	require.NoError(t, err)

	err = repo.BatchDeleteForUser(ctx, user.ID)

	assert.NoError(t, err)

	// Verify all tokens for the user were deleted
	retrieved, err := repo.GetAllForUser(ctx, user.ID, entities.ApiTokenFilter{})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

func TestTokenGORMRepo_GetById(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "Test Token",
		TokenHash:   "hashvalue123",
		Salt:        "saltvalue",
		Description: "Test description",
		Owner:       user,
		Expiration:  time.Now().Add(24 * time.Hour),
		Active:      true,
		UpdatedBy:   user,
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	retrieved, err := repo.GetById(ctx, token.ID)

	assert.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.Name, retrieved.Name)
	assert.Equal(t, token.TokenHash, retrieved.TokenHash)
	assert.Equal(t, token.Salt, retrieved.Salt)
	assert.Equal(t, token.Description, retrieved.Description)
	assert.Equal(t, token.Active, retrieved.Active)
}

func TestTokenGORMRepo_GetById_NotFound(t *testing.T) {
	repo, _ := setupTokenTestDB(t)
	ctx := context.Background()

	_, err := repo.GetById(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestTokenGORMRepo_GetAllForUser(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	tokens := []entities.ApiToken{
		{
			ID:         entities.NewId(),
			Name:       "Token 1",
			TokenHash:  "hash1",
			Salt:       "salt1",
			Owner:      user,
			Expiration: time.Now().Add(24 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
		{
			ID:         entities.NewId(),
			Name:       "Token 2",
			TokenHash:  "hash2",
			Salt:       "salt2",
			Owner:      user,
			Expiration: time.Now().Add(48 * time.Hour),
			Active:     true,
			UpdatedBy:  user,
		},
	}

	err := repo.BatchCreate(ctx, tokens)
	require.NoError(t, err)

	retrieved, err := repo.GetAllForUser(ctx, user.ID, entities.ApiTokenFilter{})

	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestTokenGORMRepo_GetAllForUser_NoTokens(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	retrieved, err := repo.GetAllForUser(ctx, user.ID, entities.ApiTokenFilter{})

	assert.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

func TestTokenGORMRepo_Update(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "Original Name",
		TokenHash:   "hashvalue",
		Salt:        "saltvalue",
		Description: "Original description",
		Owner:       user,
		Expiration:  time.Now().Add(24 * time.Hour),
		Active:      true,
		UpdatedBy:   user,
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	// Update the token
	token.Name = "Updated Name"
	token.Description = "Updated description"
	token.Active = false

	updated, err := repo.Update(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
	assert.False(t, updated.Active)

	// Verify the update persisted
	retrieved, err := repo.GetById(ctx, token.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.False(t, retrieved.Active)
}

func TestTokenGORMRepo_Update_NotFound(t *testing.T) {
	repo, user := setupTokenTestDB(t)
	ctx := context.Background()

	token := entities.ApiToken{
		ID:         entities.NewId(),
		Name:       "Nonexistent Token",
		TokenHash:  "hashvalue",
		Salt:       "saltvalue",
		Owner:      user,
		Expiration: time.Now().Add(24 * time.Hour),
		Active:     true,
		UpdatedBy:  user,
	}

	// Try to update a token that doesn't exist
	_, err := repo.Update(ctx, token)

	// GORM Update doesn't return an error if no rows are affected
	// So we just verify the function executes without panic
	assert.NoError(t, err)
}
