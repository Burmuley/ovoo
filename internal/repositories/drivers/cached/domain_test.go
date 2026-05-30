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

func TestNewCachedCustomDomainRepo_NilCache(t *testing.T) {
	rawD, err := gormdriver.NewCustomDomainGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedCustomDomainRepo(nil, rawD, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedCustomDomainRepo_NilRepo(t *testing.T) {
	repo, err := NewCachedCustomDomainRepo(newMemoryCache(t), nil, &cacheCfg)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedCustomDomainRepo_NilConfig(t *testing.T) {
	rawD, err := gormdriver.NewCustomDomainGORMRepo(newDB(t))
	require.NoError(t, err)
	repo, err := NewCachedCustomDomainRepo(newMemoryCache(t), rawD, nil)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Nil(t, repo)
}

func TestNewCachedCustomDomainRepo_Valid(t *testing.T) {
	e := setupDomainsTest(t)
	assert.NotNil(t, e.cachedDomains)
}

// --- GetById ---

func TestCustomDomainRepo_GetById_CacheMiss(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	domain := insertDomain(t, e.rawDomains, user)

	result, err := e.cachedDomains.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.ID, result.ID)
}

// CacheHit: a backstage update via rawDomains is invisible to the cached repo.
func TestCustomDomainRepo_GetById_CacheHit(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	domain := insertDomain(t, e.rawDomains, user)

	first, err := e.cachedDomains.GetById(ctx, domain.ID)
	require.NoError(t, err)

	modified := domain
	modified.Name = "modified.example.com"
	_, err = e.rawDomains.Update(ctx, modified)
	require.NoError(t, err)

	result, err := e.cachedDomains.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, first.Name, result.Name)
}

func TestCustomDomainRepo_GetById_NotFound(t *testing.T) {
	e := setupDomainsTest(t)
	_, err := e.cachedDomains.GetById(context.Background(), entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- GetAll ---

func TestCustomDomainRepo_GetAll_CacheMiss(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertDomain(t, e.rawDomains, user)
	insertDomain(t, e.rawDomains, user)

	filter := entities.CustomDomainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	result, meta, err := e.cachedDomains.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, meta.TotalRecords)
}

func TestCustomDomainRepo_GetAll_CacheHit(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertDomain(t, e.rawDomains, user)
	insertDomain(t, e.rawDomains, user)

	filter := entities.CustomDomainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedDomains.GetAll(ctx, filter)
	require.NoError(t, err)

	// Backstage insert — cache is unaware.
	insertDomain(t, e.rawDomains, user)

	result, _, err := e.cachedDomains.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Create ---

func TestCustomDomainRepo_Create_EvictsList(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	insertDomain(t, e.rawDomains, user)

	filter := entities.CustomDomainFilter{Filter: entities.Filter{Page: 1, PageSize: 10}}
	_, _, err := e.cachedDomains.GetAll(ctx, filter)
	require.NoError(t, err)

	// Create through the cached layer → evicts the list.
	insertDomain(t, e.cachedDomains, user)

	result, _, err := e.cachedDomains.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Update ---

func TestCustomDomainRepo_Update_EvictsCache(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	domain := insertDomain(t, e.rawDomains, user)

	// Populate id cache.
	_, err := e.cachedDomains.GetById(ctx, domain.ID)
	require.NoError(t, err)

	updated := domain
	updated.Name = "after-update.example.com"
	_, err = e.cachedDomains.Update(ctx, updated)
	require.NoError(t, err)

	// Id-key evicted: fresh DB value is returned.
	result, err := e.cachedDomains.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, "after-update.example.com", result.Name)
}

// --- Delete ---

func TestCustomDomainRepo_Delete_NoCachedEntry(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	domain := insertDomain(t, e.rawDomains, user)

	require.NoError(t, e.cachedDomains.Delete(ctx, user, domain.ID))

	_, err := e.cachedDomains.GetById(ctx, domain.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// WithCachedEntry: the opportunistic lookup evicts the id-key precisely.
func TestCustomDomainRepo_Delete_WithCachedEntry(t *testing.T) {
	e := setupDomainsTest(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	domain := insertDomain(t, e.rawDomains, user)

	// Populate id cache.
	_, err := e.cachedDomains.GetById(ctx, domain.ID)
	require.NoError(t, err)

	require.NoError(t, e.cachedDomains.Delete(ctx, user, domain.ID))

	_, err = e.cachedDomains.GetById(ctx, domain.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestCustomDomainRepo_Delete_NotFound(t *testing.T) {
	e := setupDomainsTest(t)
	err := e.cachedDomains.Delete(context.Background(), entities.User{}, entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
