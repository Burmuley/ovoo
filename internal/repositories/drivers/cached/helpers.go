package cached

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Burmuley/ovoo/internal/cache"
	"github.com/Burmuley/ovoo/internal/entities"
)

const (
// singleItemTTL = 5 * time.Minute
// listTTL       = 1 * time.Minute
)

func durationSeconds(d int) time.Duration {
	return time.Duration(d) * time.Second
}

// getFromCache deserializes a cached value into T. Returns (zero, false) on any miss or error.
func getFromCache[T any](ctx context.Context, c cache.Cache, key string) (T, bool) {
	var zero T
	data, err := c.Get(ctx, key)
	if err != nil {
		slog.Debug("error getting value from cache", "error", err.Error())
		return zero, false
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		slog.Debug("error unmarshaling cache value", "error", err.Error())
		return zero, false
	}
	return result, true
}

// setInCache serializes val and stores it; errors are silently ignored so callers
// always receive the value that was already fetched from the underlying repo.
func setInCache[T any](ctx context.Context, c cache.Cache, key string, val T, ttl time.Duration) {
	data, err := json.Marshal(val)
	if err != nil {
		slog.Debug("error marshaling cache value", "error", err.Error())
		return
	}
	if err := c.Set(ctx, key, data, ttl); err != nil {
		slog.Debug("error setting cache value", "error", err.Error())
	}
}

// filterKey produces a stable cache key by hashing the JSON representation of any filter.
func filterKey(prefix string, filter any) string {
	data, _ := json.Marshal(filter)
	sum := sha256.Sum256(data)
	return prefix + hex.EncodeToString(sum[:])
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// evict deletes specific cache keys; ErrNotFound is silently ignored.
func evict(ctx context.Context, c cache.Cache, keys ...string) {
	for _, k := range keys {
		_ = c.Delete(ctx, k)
	}
}

// evictPrefix deletes all keys with a given prefix; ErrNotFound is silently ignored.
func evictPrefix(ctx context.Context, c cache.Cache, prefix string) {
	_ = c.DeleteByPrefix(ctx, prefix)
}

// --- Address key builders ---

func addrIdKey(id entities.Id) string {
	return "addr:id:" + id.String()
}

func addrEmailKey(email entities.Email) string {
	return "addr:email:" + hashString(email.String())
}

func addrListPrefix() string { return "addr:list:" }

func addrListKey(filter entities.AddressFilter) string {
	return filterKey(addrListPrefix(), filter)
}

// --- Chain key builders ---

func chainHashKey(hash entities.Hash) string {
	return "chain:hash:" + hash.String()
}

func chainListPrefix() string { return "chain:list:" }

func chainListKey(filter entities.ChainFilter) string {
	return filterKey(chainListPrefix(), filter)
}

// --- Token key builders ---

func tokenIdKey(id entities.Id) string {
	return "token:id:" + id.String()
}

// tokenUserPrefix returns the cache prefix for all list entries belonging to one user.
// Keys are structured as "token:user:<userId>:<filter_hash>" so evicting by this prefix
// clears every cached list for that user without touching other users' entries.
func tokenUserPrefix(userId entities.Id) string {
	return "token:user:" + userId.String() + ":"
}

func tokenUserListKey(filter entities.ApiTokenFilter) string {
	if len(filter.UserIds) == 1 {
		return filterKey(tokenUserPrefix(filter.UserIds[0]), filter)
	}
	return filterKey("token:user:all:", filter)
}

// --- User key builders ---

func userIdKey(id entities.Id) string {
	return "user:id:" + id.String()
}

func userLoginKey(login string) string {
	return "user:login:" + hashString(login)
}

func userListPrefix() string { return "user:list:" }

func userListKey(filter entities.UserFilter) string {
	return filterKey(userListPrefix(), filter)
}
