package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAddressTestDB(t *testing.T) (*AddressGORMRepo, entities.User) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	// Create a user for addresses
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

	repo, err := NewAddressGORMRepo(db)
	require.NoError(t, err)

	return repo.(*AddressGORMRepo), user
}

func TestNewAddressGORMRepo(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	repo, err := NewAddressGORMRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewAddressGORMRepo_NilDB(t *testing.T) {
	repo, err := NewAddressGORMRepo(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.NotNil(t, repo)
}

func TestAddressGORMRepo_Create(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	address := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.AliasAddress,
		Email: entities.Email("test@example.com"),
		Owner: user,
		Metadata: entities.AddressMetadata{
			Comment:     "Test comment",
			ServiceName: "TestService",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)

	assert.NoError(t, err)

	// Verify the address was created
	retrieved, err := repo.GetById(ctx, address.ID)
	assert.NoError(t, err)
	assert.Equal(t, address.ID, retrieved.ID)
	assert.Equal(t, address.Email, retrieved.Email)
}

func TestAddressGORMRepo_Create_WithForwardAddress(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	forwardAddress := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.ProtectedAddress,
		Email:     entities.Email("forward@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, forwardAddress)
	require.NoError(t, err)

	address := entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          entities.Email("test@example.com"),
		Owner:          user,
		ForwardAddress: &forwardAddress,
		UpdatedBy:      user,
	}

	err = repo.Create(ctx, address)

	assert.NoError(t, err)

	retrieved, err := repo.GetById(ctx, address.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.ForwardAddress)
	assert.Equal(t, forwardAddress.Email, retrieved.ForwardAddress.Email)
}

func TestAddressGORMRepo_BatchCreate(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := []entities.Address{
		{
			ID:        entities.NewId(),
			Type:      entities.AliasAddress,
			Email:     entities.Email("addr1@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
		{
			ID:        entities.NewId(),
			Type:      entities.ProtectedAddress,
			Email:     entities.Email("addr2@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
	}

	err := repo.BatchCreate(ctx, addresses)

	assert.NoError(t, err)

	// Verify both addresses were created
	retrieved1, err := repo.GetById(ctx, addresses[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, addresses[0].Email, retrieved1.Email)

	retrieved2, err := repo.GetById(ctx, addresses[1].ID)
	assert.NoError(t, err)
	assert.Equal(t, addresses[1].Email, retrieved2.Email)
}

func TestAddressGORMRepo_Update(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	address := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.AliasAddress,
		Email: entities.Email("test@example.com"),
		Owner: user,
		Metadata: entities.AddressMetadata{
			Comment:     "Original comment",
			ServiceName: "OriginalService",
		},
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)
	require.NoError(t, err)

	// Update the address
	address.Metadata.Comment = "Updated comment"
	address.Metadata.ServiceName = "UpdatedService"

	err = repo.Update(ctx, address)
	assert.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetById(ctx, address.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated comment", retrieved.Metadata.Comment)
	assert.Equal(t, "UpdatedService", retrieved.Metadata.ServiceName)
}

func TestAddressGORMRepo_DeleteById(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	address := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.AliasAddress,
		Email:     entities.Email("test@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)
	require.NoError(t, err)

	err = repo.DeleteById(ctx, address.ID)
	assert.NoError(t, err)

	// Verify the address was deleted
	_, err = repo.GetById(ctx, address.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestAddressGORMRepo_DeleteById_NotFound(t *testing.T) {
	repo, _ := setupAddressTestDB(t)
	ctx := context.Background()

	err := repo.DeleteById(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestAddressGORMRepo_BatchDeleteById(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := []entities.Address{
		{
			ID:        entities.NewId(),
			Type:      entities.AliasAddress,
			Email:     entities.Email("addr1@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
		{
			ID:        entities.NewId(),
			Type:      entities.ProtectedAddress,
			Email:     entities.Email("addr2@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
	}

	err := repo.BatchCreate(ctx, addresses)
	require.NoError(t, err)

	ids := []entities.Id{addresses[0].ID, addresses[1].ID}
	err = repo.BatchDeleteById(ctx, ids)

	assert.NoError(t, err)

	// Verify both addresses were deleted
	_, err = repo.GetById(ctx, addresses[0].ID)
	assert.Error(t, err)

	_, err = repo.GetById(ctx, addresses[1].ID)
	assert.Error(t, err)
}

func TestAddressGORMRepo_GetById(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	address := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.AliasAddress,
		Email: entities.Email("test@example.com"),
		Owner: user,
		Metadata: entities.AddressMetadata{
			Comment:     "Test comment",
			ServiceName: "TestService",
		},
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)
	require.NoError(t, err)

	retrieved, err := repo.GetById(ctx, address.ID)

	assert.NoError(t, err)
	assert.Equal(t, address.ID, retrieved.ID)
	assert.Equal(t, address.Email, retrieved.Email)
	assert.Equal(t, address.Type, retrieved.Type)
	assert.Equal(t, address.Metadata.Comment, retrieved.Metadata.Comment)
}

func TestAddressGORMRepo_GetById_NotFound(t *testing.T) {
	repo, _ := setupAddressTestDB(t)
	ctx := context.Background()

	_, err := repo.GetById(ctx, entities.NewId())

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestAddressGORMRepo_GetByEmail(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	email := entities.Email("test@example.com")
	address := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.AliasAddress,
		Email:     email,
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)
	require.NoError(t, err)

	retrieved, err := repo.GetByEmail(ctx, email)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, address.Email, retrieved[0].Email)
}

func TestAddressGORMRepo_GetByEmail_NotFound(t *testing.T) {
	repo, _ := setupAddressTestDB(t)
	ctx := context.Background()

	retrieved, err := repo.GetByEmail(ctx, entities.Email("nonexistent@example.com"))

	assert.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

func TestAddressGORMRepo_GetAll(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := []entities.Address{
		{
			ID:        entities.NewId(),
			Type:      entities.AliasAddress,
			Email:     entities.Email("addr1@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
		{
			ID:        entities.NewId(),
			Type:      entities.ProtectedAddress,
			Email:     entities.Email("addr2@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
		{
			ID:        entities.NewId(),
			Type:      entities.ExternalAddress,
			Email:     entities.Email("addr3@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
	}

	err := repo.BatchCreate(ctx, addresses)
	require.NoError(t, err)

	filter := entities.AddressFilter{
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

func TestAddressGORMRepo_GetAll_FilterByType(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := []entities.Address{
		{
			ID:        entities.NewId(),
			Type:      entities.AliasAddress,
			Email:     entities.Email("alias@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
		{
			ID:        entities.NewId(),
			Type:      entities.ProtectedAddress,
			Email:     entities.Email("protected@example.com"),
			Owner:     user,
			UpdatedBy: user,
		},
	}

	err := repo.BatchCreate(ctx, addresses)
	require.NoError(t, err)

	filter := entities.AddressFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		Types: []entities.AddressType{entities.ProtectedAddress},
	}

	retrieved, metadata, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, entities.ProtectedAddress, retrieved[0].Type)
	assert.Equal(t, 1, metadata.TotalRecords)
}

func TestAddressGORMRepo_GetAll_FilterByOwner(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	address := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.AliasAddress,
		Email:     entities.Email("test@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, address)
	require.NoError(t, err)

	filter := entities.AddressFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		Owners: []entities.Id{user.ID},
	}

	retrieved, metadata, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, user.ID, retrieved[0].Owner.ID)
	assert.Equal(t, 1, metadata.TotalRecords)
}

func TestAddressGORMRepo_GetAll_FilterByServiceName(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := []entities.Address{
		{
			ID:    entities.NewId(),
			Type:  entities.AliasAddress,
			Email: entities.Email("service1@example.com"),
			Owner: user,
			Metadata: entities.AddressMetadata{
				ServiceName: "Service1",
			},
			UpdatedBy: user,
		},
		{
			ID:    entities.NewId(),
			Type:  entities.AliasAddress,
			Email: entities.Email("service2@example.com"),
			Owner: user,
			Metadata: entities.AddressMetadata{
				ServiceName: "Service2",
			},
			UpdatedBy: user,
		},
	}

	err := repo.BatchCreate(ctx, addresses)
	require.NoError(t, err)

	filter := entities.AddressFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		ServiceNames: []string{"Service1"},
	}

	retrieved, _, err := repo.GetAll(ctx, filter)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(retrieved), 1)
}

func TestAddressGORMRepo_GetAll_Pagination(t *testing.T) {
	repo, user := setupAddressTestDB(t)
	ctx := context.Background()

	addresses := make([]entities.Address, 5)
	for i := 0; i < 5; i++ {
		addresses[i] = entities.Address{
			ID:        entities.NewId(),
			Type:      entities.AliasAddress,
			Email:     entities.Email("addr" + string(rune('0'+i)) + "@example.com"),
			Owner:     user,
			UpdatedBy: user,
		}
	}

	err := repo.BatchCreate(ctx, addresses)
	require.NoError(t, err)

	filter := entities.AddressFilter{
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

func TestApplyAddressFilter_NoFilters(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	stmt := db.Model(&Address{})
	filter := entities.AddressFilter{}

	count := applyAddressFilter(stmt, filter)

	assert.NotNil(t, count)
	assert.Equal(t, int64(0), *count)
}

func TestApplyAddressFilter_WithPagination(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	stmt := db.Model(&Address{})
	filter := entities.AddressFilter{
		Filter: entities.Filter{
			Page:     2,
			PageSize: 5,
		},
	}

	count := applyAddressFilter(stmt, filter)

	assert.NotNil(t, count)
}
