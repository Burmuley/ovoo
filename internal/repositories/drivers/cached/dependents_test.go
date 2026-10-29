package cached

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories"
	gormrepo "github.com/Burmuley/ovoo/internal/repositories/drivers/gorm"
)

// sharedEnv wires cached repos over one database and one cache, as the repo factory does in
// production, so eviction across repositories can be observed.
type sharedEnv struct {
	rawUsers  repositories.UsersReadWriter
	rawTokens repositories.TokensReadWriter
	rawChains repositories.ChainReadWriter
	rawAddrs  repositories.AddressReadWriter

	users  *UsersRepo
	tokens *TokensRepo
	addrs  *AddrsRepo
	chains *ChainsRepo
}

func setupSharedEnv(t *testing.T) sharedEnv {
	t.Helper()
	db := newDB(t)
	c := newMemoryCache(t)

	rawU, err := gormrepo.NewUserGORMRepo(db)
	require.NoError(t, err)
	rawT, err := gormrepo.NewApiTokenGORMRepo(db)
	require.NoError(t, err)
	rawC, err := gormrepo.NewChainsGORMRepo(db)
	require.NoError(t, err)
	rawA, err := gormrepo.NewAddressGORMRepo(db)
	require.NoError(t, err)

	u, err := NewCachedUsersRepo(c, rawU, &cacheCfg)
	require.NoError(t, err)
	tk, err := NewCachedTokensRepo(c, rawT, &cacheCfg)
	require.NoError(t, err)
	a, err := NewCachedAddrsRepo(c, rawA, &cacheCfg)
	require.NoError(t, err)
	ch, err := NewCachedChainsRepo(c, rawC, &cacheCfg)
	require.NoError(t, err)

	return sharedEnv{rawU, rawT, rawC, rawA, u, tk, a, ch}
}

// assertTokenEvicted caches a token, removes it behind the cache's back and checks that
// it is no longer served after mutate runs.
func assertTokenEvicted(t *testing.T, mutate func(e sharedEnv, user entities.User)) {
	t.Helper()
	e := setupSharedEnv(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	tok := insertToken(t, e.rawTokens, user)

	_, err := e.tokens.GetById(ctx, tok.ID)
	require.NoError(t, err)
	require.NoError(t, e.rawTokens.Delete(ctx, user, tok.ID))

	// still served from the cache
	_, err = e.tokens.GetById(ctx, tok.ID)
	require.NoError(t, err)

	mutate(e, user)

	_, err = e.tokens.GetById(ctx, tok.ID)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func assertChainEvicted(t *testing.T, mutate func(e sharedEnv, user entities.User)) {
	t.Helper()
	e := setupSharedEnv(t)
	ctx := context.Background()
	user := insertUser(t, e.rawUsers)
	chain := insertChain(t, e.rawChains, user)

	_, err := e.chains.GetByHash(ctx, chain.Hash)
	require.NoError(t, err)
	_, err = e.rawChains.Delete(ctx, user, chain.Hash)
	require.NoError(t, err)

	// still served from the cache
	_, err = e.chains.GetByHash(ctx, chain.Hash)
	require.NoError(t, err)

	mutate(e, user)

	_, err = e.chains.GetByHash(ctx, chain.Hash)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// --- F20: user changes must not leave stale owner state in cached tokens ---

func TestUsersRepo_Update_EvictsCachedTokens(t *testing.T) {
	assertTokenEvicted(t, func(e sharedEnv, user entities.User) {
		user.Type = entities.AdminUser
		require.NoError(t, e.users.Update(context.Background(), user))
	})
}

func TestUsersRepo_Delete_EvictsCachedTokens(t *testing.T) {
	assertTokenEvicted(t, func(e sharedEnv, user entities.User) {
		admin := insertUser(t, e.rawUsers)
		require.NoError(t, e.users.Delete(context.Background(), admin, user.ID))
	})
}

// --- F24: user/address changes must not leave stale Active flags in cached chains ---

func TestUsersRepo_Update_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		user.Active = false
		require.NoError(t, e.users.Update(context.Background(), user))
	})
}

func TestUsersRepo_Delete_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		admin := insertUser(t, e.rawUsers)
		require.NoError(t, e.users.Delete(context.Background(), admin, user.ID))
	})
}

func TestAddrsRepo_Update_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		addr := insertAddress(t, e.rawAddrs, user)
		addr.Active = false
		require.NoError(t, e.addrs.Update(context.Background(), addr))
	})
}

func TestAddrsRepo_DeleteById_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		addr := insertAddress(t, e.rawAddrs, user)
		require.NoError(t, e.addrs.DeleteById(context.Background(), user, addr.ID))
	})
}

func TestAddrsRepo_BatchDeleteById_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		addr := insertAddress(t, e.rawAddrs, user)
		require.NoError(t, e.addrs.BatchDeleteById(context.Background(), user, []entities.Id{addr.ID}))
	})
}

func TestAddrsRepo_BatchUpdate_EvictsCachedChains(t *testing.T) {
	assertChainEvicted(t, func(e sharedEnv, user entities.User) {
		active := false
		require.NoError(t, e.addrs.BatchUpdate(context.Background(),
			entities.AddressFilter{Owners: []entities.Id{user.ID}},
			entities.AddressBulkUpdateFields{Active: &active}))
	})
}
