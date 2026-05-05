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

func TestNewCachedAddrsRepo_NilCache(t *testing.T) {
	rawA, err := gormdriver.NewAddressGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedAddrsRepo(nil, rawA, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedAddrsRepo_NilRepo(t *testing.T) {
	repo, err := NewCachedAddrsRepo(newMemoryCache(t), nil, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedAddrsRepo_Valid(t *testing.T) {
	e := setupAddrsTest(t)
	assert.NotNil(t, e.cachedAddrs)
}

// --- GetById ---

func TestAddrsRepo_GetById_CacheMiss(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	result, err := e.cachedAddrs.GetById(ctx, addr.ID)
	assert.NoError(t, err)
	assert.Equal(t, addr.ID, result.ID)
}

// CacheHit: after a miss the value is cached; a backstage update via rawAddrs
// is invisible to the cached repo until the TTL expires or an evicting write
// goes through the cached layer.
func TestAddrsRepo_GetById_CacheHit(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	// First call: cache miss — populates cache.
	first, err := e.cachedAddrs.GetById(ctx, addr.ID)
	require.NoError(t, err)

	// Backstage mutation via raw repo (cache is unaware).
	modified := addr
	modified.Metadata.Comment = "modified-in-db"
	require.NoError(t, e.rawAddrs.Update(ctx, modified))

	// Second call: stale cached value is returned, not the modified DB row.
	result, err := e.cachedAddrs.GetById(ctx, addr.ID)
	assert.NoError(t, err)
	assert.Equal(t, first.Metadata.Comment, result.Metadata.Comment)
}

func TestAddrsRepo_GetById_NotFound(t *testing.T) {
	e := setupAddrsTest(t)
	_, err := e.cachedAddrs.GetById(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetByEmail ---

func TestAddrsRepo_GetByEmail_CacheMiss(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	result, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// CacheHit: delete the row via rawAddrs; the cached repo still returns the
// address because it serves from the cached slice.
func TestAddrsRepo_GetByEmail_CacheHit(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	_, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	require.NoError(t, err)

	// Remove from DB without going through the cached layer.
	require.NoError(t, e.rawAddrs.DeleteById(ctx, addr.ID))

	// Cache still holds the original result.
	result, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestAddrsRepo_GetByEmail_NotFound(t *testing.T) {
	e := setupAddrsTest(t)
	result, err := e.cachedAddrs.GetByEmail(context.Background(), "noone@example.com")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// --- GetAll ---

func TestAddrsRepo_GetAll_CacheMiss(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	insertAddress(t, e.rawAddrs, owner)
	insertAddress(t, e.rawAddrs, owner)

	filter := entities.AddressFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	result, meta, err := e.cachedAddrs.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, meta.TotalRecords)
}

// CacheHit: insert a third address via rawAddrs after the list was cached;
// the cached repo still returns 2 items.
func TestAddrsRepo_GetAll_CacheHit(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	insertAddress(t, e.rawAddrs, owner)
	insertAddress(t, e.rawAddrs, owner)

	filter := entities.AddressFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedAddrs.GetAll(ctx, filter)
	require.NoError(t, err)

	// Backstage insert — cache is unaware.
	insertAddress(t, e.rawAddrs, owner)

	result, _, err := e.cachedAddrs.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Create ---

// Create via the cached layer must evict the list cache so the next GetAll
// returns the full, up-to-date result.
func TestAddrsRepo_Create_EvictsList(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	insertAddress(t, e.rawAddrs, owner)

	filter := entities.AddressFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}

	// Populate list cache (1 item).
	_, _, err := e.cachedAddrs.GetAll(ctx, filter)
	require.NoError(t, err)

	// Create through the cached layer → evicts the list.
	newAddr := insertAddress(t, e.cachedAddrs, owner)
	_ = newAddr

	// List is fresh; the new address must appear.
	result, _, err := e.cachedAddrs.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- BatchCreate ---

func TestAddrsRepo_BatchCreate_EvictsList(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	insertAddress(t, e.rawAddrs, owner)

	filter := entities.AddressFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedAddrs.GetAll(ctx, filter)
	require.NoError(t, err)

	id := entities.NewId()
	batch := []entities.Address{
		{
			ID:        id,
			Type:      entities.AliasAddress,
			Email:     entities.Email("batch-" + string(id) + "@example.com"),
			Owner:     owner,
			UpdatedBy: owner,
		},
	}
	require.NoError(t, e.cachedAddrs.BatchCreate(ctx, batch))

	result, _, err := e.cachedAddrs.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Update ---

// Update through the cached layer must evict both the id-key and the
// email-key so the next read returns the DB-fresh value.
func TestAddrsRepo_Update_EvictsCache(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	// Populate id and email caches.
	_, err := e.cachedAddrs.GetById(ctx, addr.ID)
	require.NoError(t, err)
	_, err = e.cachedAddrs.GetByEmail(ctx, addr.Email)
	require.NoError(t, err)

	// Update through the cached layer (triggers eviction of both keys).
	updated := addr
	updated.Metadata.Comment = "after-update"
	require.NoError(t, e.cachedAddrs.Update(ctx, updated))

	// Id-key evicted: fresh DB value is returned.
	result, err := e.cachedAddrs.GetById(ctx, addr.ID)
	assert.NoError(t, err)
	assert.Equal(t, "after-update", result.Metadata.Comment)

	// Email-key evicted: a backstage modification before re-read would be
	// visible — the cache is cold again after eviction.
	byEmail, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	assert.NoError(t, err)
	assert.Len(t, byEmail, 1)
	assert.Equal(t, "after-update", byEmail[0].Metadata.Comment)
}

// --- DeleteById ---

func TestAddrsRepo_DeleteById_NoCachedEntry(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	require.NoError(t, e.cachedAddrs.DeleteById(ctx, addr.ID))

	_, err := e.cachedAddrs.GetById(ctx, addr.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// WithCachedEntry: when the id-key is already cached DeleteById uses the
// opportunistic lookup to evict the email-key too, without an extra DB round-
// trip. Both caches must be cleared.
func TestAddrsRepo_DeleteById_WithCachedEntry(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	// Populate both cache keys.
	_, err := e.cachedAddrs.GetById(ctx, addr.ID)
	require.NoError(t, err)
	_, err = e.cachedAddrs.GetByEmail(ctx, addr.Email)
	require.NoError(t, err)

	require.NoError(t, e.cachedAddrs.DeleteById(ctx, addr.ID))

	// Both cache entries must be gone — DB confirms deletion.
	_, err = e.cachedAddrs.GetById(ctx, addr.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	byEmail, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	assert.NoError(t, err)
	assert.Empty(t, byEmail)
}

func TestAddrsRepo_DeleteById_NotFound(t *testing.T) {
	e := setupAddrsTest(t)
	err := e.cachedAddrs.DeleteById(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// EmailCacheStaleWhenIdNotCached: when the id-key is not in cache at delete
// time, DeleteById cannot know the email, so the email-key is intentionally
// left to expire via TTL. GetByEmail still returns the deleted address until
// the entry ages out.
func TestAddrsRepo_DeleteById_EmailCacheStaleWhenIdNotCached(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	addr := insertAddress(t, e.rawAddrs, owner)

	// Populate email cache but deliberately skip the id cache.
	_, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	require.NoError(t, err)

	// Delete through the cached layer (id-key NOT in cache — no opportunistic eviction).
	require.NoError(t, e.cachedAddrs.DeleteById(ctx, addr.ID))

	// Id-key correctly gone from DB.
	_, err = e.cachedAddrs.GetById(ctx, addr.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	// Email-key was NOT evicted — stale cached slice is still served.
	byEmail, err := e.cachedAddrs.GetByEmail(ctx, addr.Email)
	assert.NoError(t, err)
	assert.Len(t, byEmail, 1)
}

// --- BatchDeleteById ---

func TestAddrsRepo_BatchDeleteById_EvictsCache(t *testing.T) {
	e := setupAddrsTest(t)
	ctx := context.Background()
	owner := insertUser(t, e.rawUsers)
	a1 := insertAddress(t, e.rawAddrs, owner)
	a2 := insertAddress(t, e.rawAddrs, owner)

	// Populate cache entries for both addresses.
	_, err := e.cachedAddrs.GetById(ctx, a1.ID)
	require.NoError(t, err)
	_, err = e.cachedAddrs.GetById(ctx, a2.ID)
	require.NoError(t, err)

	require.NoError(t, e.cachedAddrs.BatchDeleteById(ctx, []entities.Id{a1.ID, a2.ID}))

	_, err = e.cachedAddrs.GetById(ctx, a1.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	_, err = e.cachedAddrs.GetById(ctx, a2.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
