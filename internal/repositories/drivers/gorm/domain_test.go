package gorm

import (
	"context"
	"testing"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCustomDomainTestDB(t *testing.T) (*CustomDomainGORMRepo, entities.User) {
	t.Helper()
	cfg := config.ConfigDB{
		Driver:   "gorm",
		LogLevel: "silent",
		Config: config.ConfigDBDriver{
			GORM: config.ConfigDBDriverGORM{
				Driver:           "sqlite",
				ConnectionString: ":memory:",
			},
		},
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	userRepo, err := NewUserGORMRepo(db)
	require.NoError(t, err)

	user := entities.User{
		ID:           entities.NewId(),
		Login:        "owner@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Type:         entities.RegularUser,
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(context.Background(), user))

	repo, err := NewCustomDomainGORMRepo(db)
	require.NoError(t, err)

	return repo.(*CustomDomainGORMRepo), user
}

func TestNewCustomDomainGORMRepo(t *testing.T) {
	cfg := config.ConfigDB{
		Driver:   "gorm",
		LogLevel: "silent",
		Config: config.ConfigDBDriver{
			GORM: config.ConfigDBDriverGORM{
				Driver:           "sqlite",
				ConnectionString: ":memory:",
			},
		},
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	repo, err := NewCustomDomainGORMRepo(db)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewCustomDomainGORMRepo_NilDB(t *testing.T) {
	repo, err := NewCustomDomainGORMRepo(nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.NotNil(t, repo)
}

func TestCustomDomainGORMRepo_Create(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      "example.com",
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, domain)
	assert.NoError(t, err)

	retrieved, err := repo.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.ID, retrieved.ID)
	assert.Equal(t, domain.Name, retrieved.Name)
}

func TestCustomDomainGORMRepo_Update(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      "example.com",
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	domain.Name = "updated.com"
	_, err = repo.Update(ctx, domain)
	assert.NoError(t, err)

	retrieved, err := repo.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated.com", retrieved.Name)
}

func TestCustomDomainGORMRepo_Delete(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      "example.com",
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	err = repo.Delete(ctx, user, domain.ID)
	assert.NoError(t, err)

	_, err = repo.GetById(ctx, domain.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestCustomDomainGORMRepo_Delete_NotFound(t *testing.T) {
	repo, _ := setupCustomDomainTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, entities.User{}, entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestCustomDomainGORMRepo_GetById(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      "example.com",
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	retrieved, err := repo.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.ID, retrieved.ID)
	assert.Equal(t, domain.Name, retrieved.Name)
}

func TestCustomDomainGORMRepo_GetById_NotFound(t *testing.T) {
	repo, _ := setupCustomDomainTestDB(t)
	ctx := context.Background()

	_, err := repo.GetById(ctx, entities.NewId())
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestCustomDomainGORMRepo_GetAll(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(ctx, entities.CustomDomain{
			ID:        entities.NewId(),
			Name:      "domain" + string(rune('0'+i)) + ".com",
			Owner:     user,
			UpdatedBy: user,
		}))
	}

	filter := entities.CustomDomainFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
	}

	retrieved, meta, err := repo.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
	assert.Equal(t, 3, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_FilterByOwner(t *testing.T) {
	cfg := config.ConfigDB{
		Driver:   "gorm",
		LogLevel: "silent",
		Config: config.ConfigDBDriver{
			GORM: config.ConfigDBDriverGORM{
				Driver:           "sqlite",
				ConnectionString: ":memory:",
			},
		},
	}
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	userRepo, err := NewUserGORMRepo(db)
	require.NoError(t, err)

	user1 := entities.User{ID: entities.NewId(), Login: "u1@example.com", Type: entities.RegularUser, PasswordHash: "h"}
	user2 := entities.User{ID: entities.NewId(), Login: "u2@example.com", Type: entities.RegularUser, PasswordHash: "h"}
	require.NoError(t, userRepo.Create(ctx, user1))
	require.NoError(t, userRepo.Create(ctx, user2))

	domainRepo, err := NewCustomDomainGORMRepo(db)
	require.NoError(t, err)

	require.NoError(t, domainRepo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "u1domain.com", Owner: user1, UpdatedBy: user1,
	}))
	require.NoError(t, domainRepo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "u2domain.com", Owner: user2, UpdatedBy: user2,
	}))

	filter := entities.CustomDomainFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
		Owners: []entities.Id{user1.ID},
	}

	retrieved, meta, err := domainRepo.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, user1.ID, retrieved[0].Owner.ID)
	assert.Equal(t, 1, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_FilterByActive(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	activeTrue := true
	activeFalse := false

	// Create both as active (GORM skips false zero-values on Create)
	domains := []entities.CustomDomain{
		{ID: entities.NewId(), Name: "active.com", Owner: user, UpdatedBy: user, Active: true},
		{ID: entities.NewId(), Name: "inactive.com", Owner: user, UpdatedBy: user, Active: true},
	}
	for _, d := range domains {
		require.NoError(t, repo.Create(ctx, d))
	}

	// Deactivate the second via Update (Select("*") handles zero values)
	domains[1].Active = false
	_, err := repo.Update(ctx, domains[1])
	require.NoError(t, err)

	// Filter active=true
	retrieved, meta, err := repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
		Active: &activeTrue,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.True(t, retrieved[0].Active)
	assert.Equal(t, 1, meta.TotalRecords)

	// Filter active=false
	retrieved, meta, err = repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
		Active: &activeFalse,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.False(t, retrieved[0].Active)
	assert.Equal(t, 1, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_FilterByVerified(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	verifiedTrue := true
	verifiedFalse := false

	// Create two unverified domains (Verified defaults to false)
	domains := []entities.CustomDomain{
		{ID: entities.NewId(), Name: "notverified.com", Owner: user, UpdatedBy: user},
		{ID: entities.NewId(), Name: "tobevertified.com", Owner: user, UpdatedBy: user},
	}
	for _, d := range domains {
		require.NoError(t, repo.Create(ctx, d))
	}

	// Verify the second domain via Update (Select("*") handles zero-value booleans)
	domains[1].Verified = true
	_, err := repo.Update(ctx, domains[1])
	require.NoError(t, err)

	// Filter verified=true
	retrieved, meta, err := repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:   entities.Filter{Page: 1, PageSize: 10},
		Verified: &verifiedTrue,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.True(t, retrieved[0].Verified)
	assert.Equal(t, 1, meta.TotalRecords)

	// Filter verified=false
	retrieved, meta, err = repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:   entities.Filter{Page: 1, PageSize: 10},
		Verified: &verifiedFalse,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.False(t, retrieved[0].Verified)
	assert.Equal(t, 1, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_CreateGlobal(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	domain := entities.CustomDomain{
		ID:        entities.NewId(),
		Name:      "global.example.com",
		Global:    true,
		Owner:     user,
		UpdatedBy: user,
	}

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	retrieved, err := repo.GetById(ctx, domain.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved.Global)
	assert.Equal(t, domain.Name, retrieved.Name)
}

func TestCustomDomainGORMRepo_GetAll_IncludeGlobal_WithOwner(t *testing.T) {
	cfg := config.ConfigDB{
		Driver:   "gorm",
		LogLevel: "silent",
		Config: config.ConfigDBDriver{
			GORM: config.ConfigDBDriverGORM{
				Driver:           "sqlite",
				ConnectionString: ":memory:",
			},
		},
	}
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	userRepo, err := NewUserGORMRepo(db)
	require.NoError(t, err)

	admin := entities.User{ID: entities.NewId(), Login: "admin@example.com", Type: entities.AdminUser, PasswordHash: "h"}
	user := entities.User{ID: entities.NewId(), Login: "user@example.com", Type: entities.RegularUser, PasswordHash: "h"}
	require.NoError(t, userRepo.Create(ctx, admin))
	require.NoError(t, userRepo.Create(ctx, user))

	repo, err := NewCustomDomainGORMRepo(db)
	require.NoError(t, err)

	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "global.example.com", Global: true, Owner: admin, UpdatedBy: admin,
	}))
	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "personal.example.com", Global: false, Owner: user, UpdatedBy: user,
	}))

	// IncludeGlobal=true + owner filter returns both
	retrieved, meta, err := repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:        entities.Filter{Page: 1, PageSize: 10},
		Owners:        []entities.Id{user.ID},
		IncludeGlobal: true,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, 2, meta.TotalRecords)

	// IncludeGlobal=false + same owner returns only personal
	retrieved, meta, err = repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:        entities.Filter{Page: 1, PageSize: 10},
		Owners:        []entities.Id{user.ID},
		IncludeGlobal: false,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.False(t, retrieved[0].Global)
	assert.Equal(t, 1, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_IncludeGlobal_NoOwner(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "global.example.com", Global: true, Owner: user, UpdatedBy: user,
	}))
	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "personal.example.com", Global: false, Owner: user, UpdatedBy: user,
	}))

	// IncludeGlobal=true with no Owners returns only global domains
	retrieved, meta, err := repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:        entities.Filter{Page: 1, PageSize: 10},
		IncludeGlobal: true,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.True(t, retrieved[0].Global)
	assert.Equal(t, 2, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_NotIncludeGlobal_NoOwner(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "global.example.com", Global: true, Owner: user, UpdatedBy: user,
	}))
	require.NoError(t, repo.Create(ctx, entities.CustomDomain{
		ID: entities.NewId(), Name: "personal.example.com", Global: false, Owner: user, UpdatedBy: user,
	}))

	// IncludeGlobal=true with no Owners returns only global domains
	retrieved, meta, err := repo.GetAll(ctx, entities.CustomDomainFilter{
		Filter:        entities.Filter{Page: 1, PageSize: 10},
		IncludeGlobal: false,
	})
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.False(t, retrieved[0].Global)
	assert.Equal(t, 1, meta.TotalRecords)
}

func TestCustomDomainGORMRepo_GetAll_Pagination(t *testing.T) {
	repo, user := setupCustomDomainTestDB(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(ctx, entities.CustomDomain{
			ID:        entities.NewId(),
			Name:      "domain" + string(rune('0'+i)) + ".com",
			Owner:     user,
			UpdatedBy: user,
		}))
	}

	filter := entities.CustomDomainFilter{
		Filter: entities.Filter{Page: 1, PageSize: 2},
	}

	retrieved, meta, err := repo.GetAll(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, 5, meta.TotalRecords)
	assert.Equal(t, 1, meta.CurrentPage)
	assert.Equal(t, 3, meta.LastPage)
}
