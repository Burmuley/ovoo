package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChainsTestDB(t *testing.T) (*ChainsGORMRepo, entities.User) {
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

	repo, err := NewChainsGORMRepo(db)
	require.NoError(t, err)

	return repo.(*ChainsGORMRepo), user
}

func createTestChain(user entities.User) entities.Chain {
	origFromAddr := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.ExternalAddress,
		Email:     entities.Email("origfrom@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	origToAddr := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.ExternalAddress,
		Email:     entities.Email("origto@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	fromAddr := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.AliasAddress,
		Email:     entities.Email("from@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	toAddr := entities.Address{
		ID:        entities.NewId(),
		Type:      entities.ProtectedAddress,
		Email:     entities.Email("to@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}

	hash := entities.NewHash(string(origFromAddr.Email), string(origToAddr.Email))

	return entities.Chain{
		Hash:            hash,
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
		OrigFromAddress: origFromAddr,
		OrigToAddress:   origToAddr,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UpdatedBy:       user,
	}
}

func TestNewChainsGORMRepo(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	repo, err := NewChainsGORMRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewChainsGORMRepo_NilDB(t *testing.T) {
	repo, err := NewChainsGORMRepo(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.NotNil(t, repo)
}

func TestChainsGORMRepo_Create(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain := createTestChain(user)

	err := repo.Create(ctx, chain)

	assert.NoError(t, err)

	// Verify the chain was created
	retrieved, err := repo.GetByHash(ctx, chain.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, retrieved.Hash)
	assert.Equal(t, chain.FromAddress.Email, retrieved.FromAddress.Email)
	assert.Equal(t, chain.ToAddress.Email, retrieved.ToAddress.Email)
}

func TestChainsGORMRepo_BatchCreate(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain1 := createTestChain(user)

	// Create a second chain with different emails
	chain2 := createTestChain(user)
	chain2.OrigFromAddress.Email = entities.Email("origfrom2@example.com")
	chain2.OrigToAddress.Email = entities.Email("origto2@example.com")
	chain2.Hash = entities.NewHash(string(chain2.OrigFromAddress.Email), string(chain2.OrigToAddress.Email))

	chains := []entities.Chain{chain1, chain2}

	err := repo.BatchCreate(ctx, chains)

	assert.NoError(t, err)

	// Verify both chains were created
	retrieved1, err := repo.GetByHash(ctx, chain1.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain1.Hash, retrieved1.Hash)

	retrieved2, err := repo.GetByHash(ctx, chain2.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain2.Hash, retrieved2.Hash)
}

func TestChainsGORMRepo_GetByHash(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain := createTestChain(user)

	err := repo.Create(ctx, chain)
	require.NoError(t, err)

	retrieved, err := repo.GetByHash(ctx, chain.Hash)

	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, retrieved.Hash)
	assert.Equal(t, chain.FromAddress.Email, retrieved.FromAddress.Email)
	assert.Equal(t, chain.ToAddress.Email, retrieved.ToAddress.Email)
	assert.Equal(t, chain.OrigFromAddress.Email, retrieved.OrigFromAddress.Email)
	assert.Equal(t, chain.OrigToAddress.Email, retrieved.OrigToAddress.Email)
}

func TestChainsGORMRepo_GetByHash_NotFound(t *testing.T) {
	repo, _ := setupChainsTestDB(t)
	ctx := context.Background()

	_, err := repo.GetByHash(ctx, entities.Hash("nonexistent"))

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestChainsGORMRepo_GetByFilters(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain1 := createTestChain(user)
	chain2 := createTestChain(user)
	chain2.OrigFromAddress.Email = entities.Email("origfrom2@example.com")
	chain2.OrigToAddress.Email = entities.Email("origto2@example.com")
	chain2.Hash = entities.NewHash(string(chain2.OrigFromAddress.Email), string(chain2.OrigToAddress.Email))

	err := repo.BatchCreate(ctx, []entities.Chain{chain1, chain2})
	require.NoError(t, err)

	filter := entities.ChainFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
	}

	chains, err := repo.GetByFilters(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, chains, 2)
}

func TestChainsGORMRepo_GetByFilters_FilterByOrigFromAddrIds(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain := createTestChain(user)

	err := repo.Create(ctx, chain)
	require.NoError(t, err)

	filter := entities.ChainFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		OrigFromAddrIds: []entities.Id{chain.OrigFromAddress.ID},
	}

	chains, err := repo.GetByFilters(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, chains, 1)
	assert.Equal(t, chain.Hash, chains[0].Hash)
}

func TestChainsGORMRepo_GetByFilters_FilterByOrigToAddrIds(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain := createTestChain(user)

	err := repo.Create(ctx, chain)
	require.NoError(t, err)

	filter := entities.ChainFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 10,
		},
		OrigToAddrIds: []entities.Id{chain.OrigToAddress.ID},
	}

	chains, err := repo.GetByFilters(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, chains, 1)
	assert.Equal(t, chain.Hash, chains[0].Hash)
}

func TestChainsGORMRepo_GetByFilters_Pagination(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chains := make([]entities.Chain, 3)
	for i := 0; i < 3; i++ {
		chains[i] = createTestChain(user)
		chains[i].OrigFromAddress.Email = entities.Email("origfrom" + string(rune('0'+i)) + "@example.com")
		chains[i].OrigToAddress.Email = entities.Email("origto" + string(rune('0'+i)) + "@example.com")
		chains[i].Hash = entities.NewHash(string(chains[i].OrigFromAddress.Email), string(chains[i].OrigToAddress.Email))
	}

	err := repo.BatchCreate(ctx, chains)
	require.NoError(t, err)

	filter := entities.ChainFilter{
		Filter: entities.Filter{
			Page:     1,
			PageSize: 2,
		},
	}

	retrieved, err := repo.GetByFilters(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestChainsGORMRepo_Delete(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain := createTestChain(user)

	err := repo.Create(ctx, chain)
	require.NoError(t, err)

	deleted, err := repo.Delete(ctx, chain.Hash)

	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, deleted.Hash)

	// Verify the chain was deleted
	_, err = repo.GetByHash(ctx, chain.Hash)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestChainsGORMRepo_Delete_NotFound(t *testing.T) {
	repo, _ := setupChainsTestDB(t)
	ctx := context.Background()

	_, err := repo.Delete(ctx, entities.Hash("nonexistent"))

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestChainsGORMRepo_BatchDelete(t *testing.T) {
	repo, user := setupChainsTestDB(t)
	ctx := context.Background()

	chain1 := createTestChain(user)
	chain2 := createTestChain(user)
	chain2.OrigFromAddress.Email = entities.Email("origfrom2@example.com")
	chain2.OrigToAddress.Email = entities.Email("origto2@example.com")
	chain2.Hash = entities.NewHash(string(chain2.OrigFromAddress.Email), string(chain2.OrigToAddress.Email))

	err := repo.BatchCreate(ctx, []entities.Chain{chain1, chain2})
	require.NoError(t, err)

	hashes := []entities.Hash{chain1.Hash, chain2.Hash}
	err = repo.BatchDelete(ctx, hashes)

	assert.NoError(t, err)

	// Verify both chains were deleted
	_, err = repo.GetByHash(ctx, chain1.Hash)
	assert.Error(t, err)

	_, err = repo.GetByHash(ctx, chain2.Hash)
	assert.Error(t, err)
}

func TestApplyChainFilter_NoFilters(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	stmt := db.Model(&Chain{})
	filter := entities.ChainFilter{}

	count := applyChainFilter(stmt, filter)

	assert.NotNil(t, count)
	assert.Equal(t, int64(0), *count)
}

func TestApplyChainFilter_WithPagination(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		ConnStr:  ":memory:",
		LogLevel: "silent",
	}

	db, err := NewGORMDatabase(config)
	require.NoError(t, err)

	stmt := db.Model(&Chain{})
	filter := entities.ChainFilter{
		Filter: entities.Filter{
			Page:     2,
			PageSize: 5,
		},
	}

	count := applyChainFilter(stmt, filter)

	assert.NotNil(t, count)
}
