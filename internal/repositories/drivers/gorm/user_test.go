package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestDB(t *testing.T) *UserGORMRepo {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	repo, err := NewUserGORMRepo(db)
	require.NoError(t, err)

	return repo.(*UserGORMRepo)
}

func TestNewUserGORMRepo(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	repo, err := NewUserGORMRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewUserGORMRepo_NilDB(t *testing.T) {
	repo, err := NewUserGORMRepo(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.NotNil(t, repo)
}

func TestUserGORMRepo_Create(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)

	assert.NoError(t, err)

	// Verify the user was created
	retrieved, err := repo.GetById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Login, retrieved.Login)
}

func TestUserGORMRepo_Create_Duplicate(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Try to create duplicate
	user.ID = entities.NewId() // Different ID but same login
	err = repo.Create(ctx, user)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDuplicateEntry)
}

func TestUserGORMRepo_BatchCreate(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	users := []entities.User{
		{
			ID:           entities.NewId(),
			Login:        "user1@example.com",
			FirstName:    "User",
			LastName:     "One",
			Type:         entities.RegularUser,
			PasswordHash: "hash1",
		},
		{
			ID:           entities.NewId(),
			Login:        "user2@example.com",
			FirstName:    "User",
			LastName:     "Two",
			Type:         entities.AdminUser,
			PasswordHash: "hash2",
		},
	}

	err := repo.BatchCreate(ctx, users)

	assert.NoError(t, err)

	// Verify both users were created
	retrieved1, err := repo.GetByLogin(ctx, "user1@example.com")
	assert.NoError(t, err)
	assert.Equal(t, users[0].ID, retrieved1.ID)

	retrieved2, err := repo.GetByLogin(ctx, "user2@example.com")
	assert.NoError(t, err)
	assert.Equal(t, users[1].ID, retrieved2.ID)
}

func TestUserGORMRepo_Update(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update the user
	user.FirstName = "Jane"
	user.LastName = "Smith"

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", retrieved.FirstName)
	assert.Equal(t, "Smith", retrieved.LastName)
}

func TestUserGORMRepo_Delete(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify the user was deleted
	_, err = repo.GetById(ctx, user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestUserGORMRepo_Delete_NotFound(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestUserGORMRepo_GetById(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetById(ctx, user.ID)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Login, retrieved.Login)
	assert.Equal(t, user.FirstName, retrieved.FirstName)
}

func TestUserGORMRepo_GetById_NotFound(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	_, err := repo.GetById(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestUserGORMRepo_GetByLogin(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	user := entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByLogin(ctx, "john.doe@example.com")

	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Login, retrieved.Login)
}

func TestUserGORMRepo_GetByLogin_NotFound(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	_, err := repo.GetByLogin(ctx, "nonexistent@example.com")

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestUserGORMRepo_GetAll(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	users := []entities.User{
		{
			ID:           entities.NewId(),
			Login:        "user1@example.com",
			FirstName:    "User",
			LastName:     "One",
			Type:         entities.RegularUser,
			PasswordHash: "hash1",
		},
		{
			ID:           entities.NewId(),
			Login:        "user2@example.com",
			FirstName:    "User",
			LastName:     "Two",
			Type:         entities.AdminUser,
			PasswordHash: "hash2",
		},
		{
			ID:           entities.NewId(),
			Login:        "user3@example.com",
			FirstName:    "User",
			LastName:     "Three",
			Type:         entities.RegularUser,
			PasswordHash: "hash3",
		},
	}

	err := repo.BatchCreate(ctx, users)
	require.NoError(t, err)

	filter := entities.UserFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
	}

	retrieved, metadata, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
	assert.Equal(t, 3, metadata.TotalRecords)
}

func TestUserGORMRepo_GetAll_WithFilter(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	users := []entities.User{
		{
			ID:           entities.NewId(),
			Login:        "user1@example.com",
			Type:         entities.RegularUser,
			PasswordHash: "hash1",
		},
		{
			ID:           entities.NewId(),
			Login:        "admin@example.com",
			Type:         entities.AdminUser,
			PasswordHash: "hash2",
		},
	}

	err := repo.BatchCreate(ctx, users)
	require.NoError(t, err)

	filter := entities.UserFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		Types: []entities.UserType{entities.AdminUser},
	}

	retrieved, metadata, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, entities.AdminUser, retrieved[0].Type)
	assert.Equal(t, 1, metadata.TotalRecords)
}

func TestUserGORMRepo_GetAll_Pagination(t *testing.T) {
	repo := setupUserTestDB(t)
	ctx := context.Background()

	users := make([]entities.User, 5)
	for i := 0; i < 5; i++ {
		users[i] = entities.User{
			ID:           entities.NewId(),
			Login:        "user" + string(rune('0'+i)) + "@example.com",
			Type:         entities.RegularUser,
			PasswordHash: "hash",
		}
	}

	err := repo.BatchCreate(ctx, users)
	require.NoError(t, err)

	filter := entities.UserFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 2,
		},
	}

	retrieved, metadata, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, 5, metadata.TotalRecords)
	assert.Equal(t, 1, metadata.CurrentPage)
	assert.Equal(t, 3, metadata.LastPage)
}
