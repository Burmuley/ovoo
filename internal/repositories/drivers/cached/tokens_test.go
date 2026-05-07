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

func TestNewCachedTokensRepo_NilCache(t *testing.T) {
	rawT, err := gormdriver.NewApiTokenGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedTokensRepo(nil, rawT, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedTokensRepo_NilRepo(t *testing.T) {
	repo, err := NewCachedTokensRepo(newMemoryCache(t), nil, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedTokensRepo_Valid(t *testing.T) {
	e := setupTokensTest(t)
	assert.NotNil(t, e.cachedTokens)
}

// --- GetById ---

func TestTokensRepo_GetById_CacheMiss(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	token := insertToken(t, e.rawTokens, user)

	result, err := e.cachedTokens.GetById(ctx, token.ID)
	assert.NoError(t, err)
	assert.Equal(t, token.ID, result.ID)
}

// CacheHit: deactivate the token via rawTokens; the cached repo still returns
// the original active=true value because it serves from cache.
func TestTokensRepo_GetById_CacheHit(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	token := insertToken(t, e.rawTokens, user)

	_, err := e.cachedTokens.GetById(ctx, token.ID)
	require.NoError(t, err)

	// Backstage mutation via raw repo.
	deactivated := token
	deactivated.Active = false
	_, err = e.rawTokens.Update(ctx, deactivated)
	require.NoError(t, err)

	result, err := e.cachedTokens.GetById(ctx, token.ID)
	assert.NoError(t, err)
	assert.True(t, result.Active) // stale cached value
}

func TestTokensRepo_GetById_NotFound(t *testing.T) {
	e := setupTokensTest(t)
	_, err := e.cachedTokens.GetById(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetAllForUser ---

func TestTokensRepo_GetAllForUser_CacheMiss(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user)
	insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// CacheHit: a backstage insert is not seen because the list is cached.
func TestTokensRepo_GetAllForUser_CacheHit(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err := e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	// Backstage insert — cache is unaware.
	insertToken(t, e.rawTokens, user)

	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// --- Create ---

// Create through the cached layer must evict the user-scoped list prefix so
// the next GetAllForUser returns the updated count.
func TestTokensRepo_Create_EvictsUserList(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err := e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	// Create through the cached layer → evicts the user-prefix cache.
	insertToken(t, e.cachedTokens, user)

	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Update ---

func TestTokensRepo_Update_EvictsCache(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	token := insertToken(t, e.rawTokens, user)

	// Populate id-key and user-list-key.
	_, err := e.cachedTokens.GetById(ctx, token.ID)
	require.NoError(t, err)
	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err = e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	updated := token
	updated.Active = false
	result, err := e.cachedTokens.Update(ctx, updated)
	assert.NoError(t, err)
	assert.False(t, result.Active)

	// Id-key evicted: fresh DB value is returned.
	fresh, err := e.cachedTokens.GetById(ctx, token.ID)
	assert.NoError(t, err)
	assert.False(t, fresh.Active)

	// User-list evicted: the cached layer no longer returns the stale [token(active=true)]
	// slice. Fresh DB data is returned — 1 token with Active=false proves the cache was
	// cleared and we are not seeing the old stale value.
	list, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.False(t, list[0].Active)
}

// --- BatchCreate ---

func TestTokensRepo_BatchCreate_EvictsUserList(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err := e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	id := entities.NewId()
	batch := []entities.ApiToken{
		{ID: id, Name: "batch-" + string(id), TokenHash: "h-" + string(id), Salt: "s", Owner: user, Active: true, UpdatedBy: user},
	}
	require.NoError(t, e.cachedTokens.BatchCreate(ctx, batch))

	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// BatchCreate with tokens belonging to different owners must evict each
// owner's user-prefix independently.
func TestTokensRepo_BatchCreate_EvictsPerOwner(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user1 := insertUser(t, e.rawUsers)
	user2 := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user1)
	insertToken(t, e.rawTokens, user2)

	filter1 := entities.ApiTokenFilter{UserIds: []entities.Id{user1.ID}}
	filter2 := entities.ApiTokenFilter{UserIds: []entities.Id{user2.ID}}

	_, err := e.cachedTokens.GetAllForUser(ctx, filter1)
	require.NoError(t, err)
	_, err = e.cachedTokens.GetAllForUser(ctx, filter2)
	require.NoError(t, err)

	id1, id2 := entities.NewId(), entities.NewId()
	batch := []entities.ApiToken{
		{ID: id1, Name: "b1-" + string(id1), TokenHash: "h-" + string(id1), Salt: "s", Owner: user1, Active: true, UpdatedBy: user1},
		{ID: id2, Name: "b2-" + string(id2), TokenHash: "h-" + string(id2), Salt: "s", Owner: user2, Active: true, UpdatedBy: user2},
	}
	require.NoError(t, e.cachedTokens.BatchCreate(ctx, batch))

	// Both user caches must be evicted — each user now has 2 tokens.
	r1, err := e.cachedTokens.GetAllForUser(ctx, filter1)
	assert.NoError(t, err)
	assert.Len(t, r1, 2)

	r2, err := e.cachedTokens.GetAllForUser(ctx, filter2)
	assert.NoError(t, err)
	assert.Len(t, r2, 2)
}

// --- Delete ---

func TestTokensRepo_Delete_NoCachedToken(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	token := insertToken(t, e.rawTokens, user)

	// Populate user-list cache before deleting (id-key is NOT cached).
	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err := e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	require.NoError(t, e.cachedTokens.Delete(ctx, token.ID))

	// Id-key miss confirms deletion.
	_, err = e.cachedTokens.GetById(ctx, token.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	// Broad "token:user:" eviction clears the list even without the id-key cached.
	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// WithCachedToken: the opportunistic lookup uses the cached owner info to
// evict only that user's list prefix (rather than all token:user: keys).
func TestTokensRepo_Delete_WithCachedToken(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	token := insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}

	// Populate both the id-key and the user-list-key.
	_, err := e.cachedTokens.GetById(ctx, token.ID)
	require.NoError(t, err)
	_, err = e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	require.NoError(t, e.cachedTokens.Delete(ctx, token.ID))

	// Both caches must be gone.
	_, err = e.cachedTokens.GetById(ctx, token.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestTokensRepo_Delete_NotFound(t *testing.T) {
	e := setupTokensTest(t)
	err := e.cachedTokens.Delete(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- BatchDeleteById ---

func TestTokensRepo_BatchDeleteById_EvictsAll(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	t1 := insertToken(t, e.rawTokens, user)
	t2 := insertToken(t, e.rawTokens, user)

	// Populate id-key caches and the user-list cache.
	_, err := e.cachedTokens.GetById(ctx, t1.ID)
	require.NoError(t, err)
	_, err = e.cachedTokens.GetById(ctx, t2.ID)
	require.NoError(t, err)
	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err = e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	require.NoError(t, e.cachedTokens.BatchDeleteById(ctx, []entities.Id{t1.ID, t2.ID}))

	_, err = e.cachedTokens.GetById(ctx, t1.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	_, err = e.cachedTokens.GetById(ctx, t2.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	// "token:" prefix eviction covers user-list keys too.
	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// --- BatchDeleteForUser ---

func TestTokensRepo_BatchDeleteForUser_EvictsUserList(t *testing.T) {
	e := setupTokensTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertToken(t, e.rawTokens, user)
	insertToken(t, e.rawTokens, user)

	filter := entities.ApiTokenFilter{UserIds: []entities.Id{user.ID}}
	_, err := e.cachedTokens.GetAllForUser(ctx, filter)
	require.NoError(t, err)

	require.NoError(t, e.cachedTokens.BatchDeleteForUser(ctx, user.ID))

	// Cache evicted; DB now returns empty list.
	result, err := e.cachedTokens.GetAllForUser(ctx, filter)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
