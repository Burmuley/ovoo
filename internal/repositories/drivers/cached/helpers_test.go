package cached

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Burmuley/ovoo/internal/cache/drivers/memory"
	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	gormrepo "github.com/Burmuley/ovoo/internal/repositories/drivers/gorm"
)

// sqliteCfg is the shared in-memory SQLite config used by all test setups.
var sqliteCfg = config.ConfigDB{
	Driver:   "gorm",
	LogLevel: "silent",
	Config: config.ConfigDBDriver{
		GORM: config.ConfigDBDriverGORM{
			Driver:           "sqlite",
			ConnectionString: ":memory:",
		},
	},
}

var cacheCfg = config.ConfigCache{
	SingleItemTTL: 300,
	ListTTL:       60,
}

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormrepo.NewDatabase(sqliteCfg)
	require.NoError(t, err)
	return db
}

func newMemoryCache(t *testing.T) *memory.MemoryCache {
	t.Helper()
	c, err := memory.New()
	require.NoError(t, err)
	return c
}

// --- Per-entity test environments ---
//
// Each environment holds a raw (uncached) repo and a cached repo that both
// point at the same in-memory SQLite database. The raw repo is used to seed
// data and perform "backstage" mutations that the cache layer is unaware of,
// which lets tests prove cache-hit / eviction behaviour through observable
// outcomes rather than call-count assertions.

type addrsTestEnv struct {
	rawUsers    repositories.UsersReadWriter
	rawAddrs    repositories.AddressReadWriter
	cachedAddrs *AddrsRepo
}

func setupAddrsTest(t *testing.T) addrsTestEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)
	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	rawA, err := gormrepo.NewAddressGORMRepo(db)
	require.NoError(t, err)
	ca, err := NewCachedAddrsRepo(c, rawA, &cacheCfg)
	require.NoError(t, err)
	return addrsTestEnv{rawU, rawA, ca}
}

type usersTestEnv struct {
	rawUsers    repositories.UsersReadWriter
	cachedUsers *UsersRepo
}

func setupUsersTest(t *testing.T) usersTestEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)
	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	cu, err := NewCachedUsersRepo(c, rawU, &cacheCfg)
	require.NoError(t, err)
	return usersTestEnv{rawU, cu}
}

type chainsTestEnv struct {
	rawUsers     repositories.UsersReadWriter
	rawChains    repositories.ChainReadWriter
	cachedChains *ChainsRepo
}

func setupChainsTest(t *testing.T) chainsTestEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)
	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	rawC, err := gormrepo.NewChainsGORMRepo(db)
	require.NoError(t, err)
	cc, err := NewCachedChainsRepo(c, rawC, &cacheCfg)
	require.NoError(t, err)
	return chainsTestEnv{rawU, rawC, cc}
}

type tokensTestEnv struct {
	rawUsers     repositories.UsersReadWriter
	rawTokens    repositories.TokensReadWriter
	cachedTokens *TokensRepo
}

func setupTokensTest(t *testing.T) tokensTestEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)
	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	rawT, err := gormrepo.NewApiTokenGORMRepo(db)
	require.NoError(t, err)
	ct, err := NewCachedTokensRepo(c, rawT, &cacheCfg)
	require.NoError(t, err)
	return tokensTestEnv{rawU, rawT, ct}
}

// --- Data-insertion helpers ---

func insertUser(t *testing.T, repo repositories.UsersReadWriter) entities.User {
	t.Helper()
	id := entities.NewId()
	u := entities.User{
		ID:           id,
		Login:        "user-" + string(id) + "@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Type:         entities.RegularUser,
		PasswordHash: "hash",
	}
	require.NoError(t, repo.Create(context.Background(), u))
	return u
}

func insertAddress(t *testing.T, repo repositories.AddressReadWriter, owner entities.User) entities.Address {
	t.Helper()
	id := entities.NewId()
	a := entities.Address{
		ID:        id,
		Type:      entities.AliasAddress,
		Email:     entities.Email("addr-" + string(id) + "@example.com"),
		Owner:     owner,
		UpdatedBy: owner,
	}
	require.NoError(t, repo.Create(context.Background(), a))
	return a
}

// insertChain builds a chain with four unique embedded addresses (no separate
// address table rows are required) and inserts it via repo.
func insertChain(t *testing.T, repo repositories.ChainReadWriter, user entities.User) entities.Chain {
	t.Helper()
	id1, id2, id3, id4 := entities.NewId(), entities.NewId(), entities.NewId(), entities.NewId()
	origFrom := entities.Address{
		ID:        id1,
		Type:      entities.ExternalAddress,
		Email:     entities.Email("origfrom-" + string(id1) + "@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}
	origTo := entities.Address{
		ID:        id2,
		Type:      entities.ExternalAddress,
		Email:     entities.Email("origto-" + string(id2) + "@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}
	from := entities.Address{
		ID:        id3,
		Type:      entities.AliasAddress,
		Email:     entities.Email("from-" + string(id3) + "@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}
	to := entities.Address{
		ID:        id4,
		Type:      entities.ProtectedAddress,
		Email:     entities.Email("to-" + string(id4) + "@example.com"),
		Owner:     user,
		UpdatedBy: user,
	}
	chain := entities.Chain{
		Hash:            entities.NewHash(string(origFrom.Email), string(origTo.Email)),
		FromAddress:     from,
		ToAddress:       to,
		OrigFromAddress: origFrom,
		OrigToAddress:   origTo,
		UpdatedBy:       user,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), chain))
	return chain
}

type domainsTestEnv struct {
	rawUsers      repositories.UsersReadWriter
	rawDomains    repositories.CustomDomainsReadWriter
	cachedDomains *CustomDomainRepo
}

func setupDomainsTest(t *testing.T) domainsTestEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)
	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	rawD, err := gormrepo.NewCustomDomainGORMRepo(db)
	require.NoError(t, err)
	cd, err := NewCachedCustomDomainRepo(c, rawD, &cacheCfg)
	require.NoError(t, err)
	return domainsTestEnv{rawU, rawD, cd}
}

func insertDomain(t *testing.T, repo repositories.CustomDomainsReadWriter, owner entities.User) entities.CustomDomain {
	t.Helper()
	id := entities.NewId()
	d := entities.CustomDomain{
		ID:        id,
		Name:      "domain-" + string(id) + ".example.com",
		Owner:     owner,
		Active:    true,
		UpdatedBy: owner,
	}
	require.NoError(t, repo.Create(context.Background(), d))
	return d
}

func insertToken(t *testing.T, repo repositories.TokensReadWriter, owner entities.User) entities.ApiToken {
	t.Helper()
	id := entities.NewId()
	tok := entities.ApiToken{
		ID:         id,
		Name:       "token-" + string(id),
		TokenHash:  "hash-" + string(id),
		Salt:       "salt",
		Owner:      owner,
		Active:     true,
		Expiration: time.Now().Add(time.Hour),
		UpdatedBy:  owner,
	}
	require.NoError(t, repo.Create(context.Background(), tok))
	return tok
}
