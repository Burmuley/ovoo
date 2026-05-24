package cached

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
	gormdriver "github.com/Burmuley/ovoo/internal/repositories/drivers/gorm"
)

// --- Constructor ---

func TestNewCachedChainsRepo_NilCache(t *testing.T) {
	rawC, err := gormdriver.NewChainsGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedChainsRepo(nil, rawC, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedChainsRepo_NilRepo(t *testing.T) {
	repo, err := NewCachedChainsRepo(newMemoryCache(t), nil, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedChainsRepo_Valid(t *testing.T) {
	e := setupChainsTest(t)
	assert.NotNil(t, e.cachedChains)
}

// --- GetByHash ---

func TestChainsRepo_GetByHash_CacheMiss(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	chain := insertChain(t, e.rawChains, user)

	result, err := e.cachedChains.GetByHash(ctx, chain.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, result.Hash)
}

// CacheHit: delete the row via rawChains — the cached repo still returns the
// chain because it serves from cache.
func TestChainsRepo_GetByHash_CacheHit(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	chain := insertChain(t, e.rawChains, user)

	_, err := e.cachedChains.GetByHash(ctx, chain.Hash)
	require.NoError(t, err)

	// Remove from DB without going through the cached layer.
	_, err = e.rawChains.Delete(ctx, user, chain.Hash)
	require.NoError(t, err)

	// Cache still holds the chain.
	result, err := e.cachedChains.GetByHash(ctx, chain.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, result.Hash)
}

func TestChainsRepo_GetByHash_NotFound(t *testing.T) {
	e := setupChainsTest(t)
	_, err := e.cachedChains.GetByHash(context.Background(), entities.NewHash("a@x.com", "b@x.com"))
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetByFilters ---

func TestChainsRepo_GetByFilters_CacheMiss(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertChain(t, e.rawChains, user)
	insertChain(t, e.rawChains, user)

	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// CacheHit: a backstage insert is not seen because the list is cached.
func TestChainsRepo_GetByFilters_CacheHit(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertChain(t, e.rawChains, user)

	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, err := e.cachedChains.GetByFilters(ctx, filter)
	require.NoError(t, err)

	// Backstage insert — cache unaware.
	insertChain(t, e.rawChains, user)

	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// --- Create ---

func TestChainsRepo_Create_EvictsList(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertChain(t, e.rawChains, user)

	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, err := e.cachedChains.GetByFilters(ctx, filter)
	require.NoError(t, err)

	// Create through the cached layer → evicts the list.
	insertChain(t, e.cachedChains, user)

	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- BatchCreate ---

func TestChainsRepo_BatchCreate_EvictsAll(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	existing := insertChain(t, e.rawChains, user)

	// Populate both a hash-key and a list-key.
	_, err := e.cachedChains.GetByHash(ctx, existing.Hash)
	require.NoError(t, err)
	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, err = e.cachedChains.GetByFilters(ctx, filter)
	require.NoError(t, err)

	// Build a new chain to batch-create.
	newChain := insertChain(t, e.rawChains, user)          // insert via raw so we have the struct
	_, err = e.rawChains.Delete(ctx, user, newChain.Hash) // remove it so we can re-create via cached
	require.NoError(t, err)
	require.NoError(t, e.cachedChains.BatchCreate(ctx, []entities.Chain{newChain}))

	// Both caches evicted: list now reflects two items, hash key miss on DB.
	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Delete ---

func TestChainsRepo_Delete_Success(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	chain := insertChain(t, e.rawChains, user)

	// Populate both hash cache and list cache.
	_, err := e.cachedChains.GetByHash(ctx, chain.Hash)
	require.NoError(t, err)
	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, err = e.cachedChains.GetByFilters(ctx, filter)
	require.NoError(t, err)

	deleted, err := e.cachedChains.Delete(ctx, user, chain.Hash)
	assert.NoError(t, err)
	assert.Equal(t, chain.Hash, deleted.Hash)

	// Hash-key evicted: DB confirms the chain is gone.
	_, err = e.cachedChains.GetByHash(ctx, chain.Hash)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	// List-key evicted: the deleted chain no longer appears in the list.
	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestChainsRepo_Delete_NotFound(t *testing.T) {
	e := setupChainsTest(t)
	_, err := e.cachedChains.Delete(context.Background(), entities.User{}, entities.NewHash("a@x.com", "b@x.com"))
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- BatchDelete ---

func TestChainsRepo_BatchDelete_EvictsAll(t *testing.T) {
	e := setupChainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	c1 := insertChain(t, e.rawChains, user)
	c2 := insertChain(t, e.rawChains, user)

	// Populate caches.
	_, err := e.cachedChains.GetByHash(ctx, c1.Hash)
	require.NoError(t, err)
	filter := entities.ChainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, err = e.cachedChains.GetByFilters(ctx, filter)
	require.NoError(t, err)

	require.NoError(t, e.cachedChains.BatchDelete(ctx, user, []entities.Hash{c1.Hash, c2.Hash}))

	// All caches evicted.
	_, err = e.cachedChains.GetByHash(ctx, c1.Hash)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	_, err = e.cachedChains.GetByHash(ctx, c2.Hash)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	result, err := e.cachedChains.GetByFilters(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
