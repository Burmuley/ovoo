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

func TestNewCachedUsersRepo_NilCache(t *testing.T) {
	rawU, err := gormdriver.NewUserGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedUsersRepo(nil, rawU, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedUsersRepo_NilRepo(t *testing.T) {
	repo, err := NewCachedUsersRepo(newMemoryCache(t), nil, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedUsersRepo_Valid(t *testing.T) {
	e := setupUsersTest(t)
	assert.NotNil(t, e.cachedUsers)
}

// --- GetById ---

func TestUsersRepo_GetById_CacheMiss(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	result, err := e.cachedUsers.GetById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
}

// CacheHit: a backstage update via rawUsers is invisible to the cached repo.
func TestUsersRepo_GetById_CacheHit(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	first, err := e.cachedUsers.GetById(ctx, user.ID)
	require.NoError(t, err)

	modified := user
	modified.FirstName = "modified-in-db"
	require.NoError(t, e.rawUsers.Update(ctx, modified))

	result, err := e.cachedUsers.GetById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, first.FirstName, result.FirstName)
}

func TestUsersRepo_GetById_NotFound(t *testing.T) {
	e := setupUsersTest(t)
	_, err := e.cachedUsers.GetById(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetByLogin ---

func TestUsersRepo_GetByLogin_CacheMiss(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	result, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	assert.NoError(t, err)
	assert.Equal(t, user.Login, result.Login)
}

func TestUsersRepo_GetByLogin_CacheHit(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	first, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	require.NoError(t, err)

	modified := user
	modified.FirstName = "modified-in-db"
	require.NoError(t, e.rawUsers.Update(ctx, modified))

	result, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	assert.NoError(t, err)
	assert.Equal(t, first.FirstName, result.FirstName)
}

func TestUsersRepo_GetByLogin_NotFound(t *testing.T) {
	e := setupUsersTest(t)
	_, err := e.cachedUsers.GetByLogin(context.Background(), "noone@example.com")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetAll ---

func TestUsersRepo_GetAll_CacheMiss(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	insertUser(t, e.rawUsers)
	insertUser(t, e.rawUsers)

	filter := entities.UserFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	result, meta, err := e.cachedUsers.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, meta.TotalRecords)
}

func TestUsersRepo_GetAll_CacheHit(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	insertUser(t, e.rawUsers)
	insertUser(t, e.rawUsers)

	filter := entities.UserFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedUsers.GetAll(ctx, filter)
	require.NoError(t, err)

	// Backstage insert — cache is unaware.
	insertUser(t, e.rawUsers)

	result, _, err := e.cachedUsers.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Create ---

func TestUsersRepo_Create_EvictsList(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	insertUser(t, e.rawUsers)

	filter := entities.UserFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedUsers.GetAll(ctx, filter)
	require.NoError(t, err)

	// Create through the cached layer → evicts the list.
	insertUser(t, e.cachedUsers)

	result, _, err := e.cachedUsers.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- BatchCreate ---

func TestUsersRepo_BatchCreate_EvictsList(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	insertUser(t, e.rawUsers)

	filter := entities.UserFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedUsers.GetAll(ctx, filter)
	require.NoError(t, err)

	id := entities.NewId()
	require.NoError(t, e.cachedUsers.BatchCreate(ctx, []entities.User{
		{ID: id, Login: "batch-" + string(id) + "@example.com", Type: entities.RegularUser, PasswordHash: "h"},
	}))

	result, _, err := e.cachedUsers.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Update ---

func TestUsersRepo_Update_EvictsCache(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	// Populate both id and login caches.
	_, err := e.cachedUsers.GetById(ctx, user.ID)
	require.NoError(t, err)
	_, err = e.cachedUsers.GetByLogin(ctx, user.Login)
	require.NoError(t, err)

	updated := user
	updated.FirstName = "after-update"
	require.NoError(t, e.cachedUsers.Update(ctx, updated))

	// Id-key evicted: fresh DB value is returned.
	result, err := e.cachedUsers.GetById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "after-update", result.FirstName)

	// Login-key evicted: fresh DB value is returned.
	byLogin, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	assert.NoError(t, err)
	assert.Equal(t, "after-update", byLogin.FirstName)
}

// --- Delete ---

func TestUsersRepo_Delete_NoCachedEntry(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	require.NoError(t, e.cachedUsers.Delete(ctx, user, user.ID))

	_, err := e.cachedUsers.GetById(ctx, user.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// WithCachedEntry: the opportunistic lookup evicts both the id-key and the
// login-key without an extra DB round-trip.
func TestUsersRepo_Delete_WithCachedEntry(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	// Populate both cache keys.
	_, err := e.cachedUsers.GetById(ctx, user.ID)
	require.NoError(t, err)
	_, err = e.cachedUsers.GetByLogin(ctx, user.Login)
	require.NoError(t, err)

	require.NoError(t, e.cachedUsers.Delete(ctx, user, user.ID))

	_, err = e.cachedUsers.GetById(ctx, user.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	_, err = e.cachedUsers.GetByLogin(ctx, user.Login)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestUsersRepo_Delete_NotFound(t *testing.T) {
	e := setupUsersTest(t)
	err := e.cachedUsers.Delete(context.Background(), entities.User{}, entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// LoginCacheStaleWhenIdNotCached: when the id-key is not in cache at delete
// time, Delete cannot know the login, so the login-key is intentionally left
// to expire via TTL. GetByLogin still returns the deleted user until the entry
// ages out.
func TestUsersRepo_Delete_LoginCacheStaleWhenIdNotCached(t *testing.T) {
	e := setupUsersTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)

	// Populate login cache but deliberately skip the id cache.
	_, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	require.NoError(t, err)

	// Delete through the cached layer (id-key NOT in cache — no opportunistic eviction).
	require.NoError(t, e.cachedUsers.Delete(ctx, user, user.ID))

	// Id-key correctly gone from DB.
	_, err = e.cachedUsers.GetById(ctx, user.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	// Login-key was NOT evicted — stale cached value is still served.
	byLogin, err := e.cachedUsers.GetByLogin(ctx, user.Login)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, byLogin.ID)
}
