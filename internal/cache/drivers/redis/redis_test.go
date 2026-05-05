package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestCache(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	addr := mr.Addr()
	cfg := config.APICacheConfig{
		Config: config.APICacheDriverConfig{
			Redis: &config.APICacheDriverRedisConfig{
				Addr: &addr,
				DB:   0,
				// Protocol 0 → New() falls back to default (3); miniredis supports HELLO.
			},
		},
	}
	c, err := New(cfg)
	require.NoError(t, err)
	return c, mr
}

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

const longTTL = time.Hour

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_ValidConfig_ReturnsNonNilCache(t *testing.T) {
	c, _ := newTestCache(t)
	assert.NotNil(t, c)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestGet_ExistingKey(t *testing.T) {
	c, mr := newTestCache(t)
	mr.Set("k", "hello")

	got, err := c.Get(context.Background(), "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)
}

func TestGet_MissingKey_ReturnsRedisNil(t *testing.T) {
	c, _ := newTestCache(t)

	_, err := c.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestGet_EmptyCache_ReturnsRedisNil(t *testing.T) {
	c, _ := newTestCache(t)

	_, err := c.Get(context.Background(), "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Set → Get round-trips
// ---------------------------------------------------------------------------

func TestSet_StoresAndRetrieves(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("hello"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)
}

func TestSet_OverwriteLastWriteWins(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("first"), longTTL))
	require.NoError(t, c.Set(ctx, "k", []byte("second"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), got)
}

func TestSet_OverwriteShortensTTL(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v1"), longTTL))
	require.NoError(t, c.Set(ctx, "k", []byte("v2"), time.Second))

	mr.FastForward(2 * time.Second)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestSet_OverwriteExtendsExpiredTTL(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v1"), time.Second))
	mr.FastForward(2 * time.Second)

	require.NoError(t, c.Set(ctx, "k", []byte("v2"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), got)
}

func TestSet_EmptyByteSliceValue(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte{}, longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Empty(t, got)
}

// ---------------------------------------------------------------------------
// TTL expiry
// ---------------------------------------------------------------------------

func TestSet_TTLExpiry_KeyUnavailable(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), time.Second))
	mr.FastForward(2 * time.Second)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestSet_LargeTTL_KeyStillAvailable(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), 24*365*time.Hour))
	mr.FastForward(time.Hour)

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("v"), got)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_ExistingKey(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), longTTL))
	require.NoError(t, c.Delete(ctx, "k"))

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// Redis DEL on a missing key returns 0 (not an error). The driver propagates this as nil.
func TestDelete_MissingKey_ReturnsNil(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.Delete(context.Background(), "missing")
	assert.NoError(t, err)
}

func TestDelete_EmptyCache_ReturnsNil(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.Delete(context.Background(), "k")
	assert.NoError(t, err)
}

func TestDelete_AlreadyDeleted_ReturnsNil(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), longTTL))
	require.NoError(t, c.Delete(ctx, "k"))

	// Second delete on missing key is still nil.
	assert.NoError(t, c.Delete(ctx, "k"))
}

// ---------------------------------------------------------------------------
// DeleteByPrefix
// ---------------------------------------------------------------------------

func TestDeleteByPrefix_MatchingKeys(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "user:1", []byte("a"), longTTL))
	require.NoError(t, c.Set(ctx, "user:2", []byte("b"), longTTL))
	require.NoError(t, c.Set(ctx, "token:1", []byte("c"), longTTL))

	require.NoError(t, c.DeleteByPrefix(ctx, "user:"))

	_, err1 := c.Get(ctx, "user:1")
	_, err2 := c.Get(ctx, "user:2")
	assert.ErrorIs(t, err1, entities.ErrNotFound)
	assert.ErrorIs(t, err2, entities.ErrNotFound)

	got, err := c.Get(ctx, "token:1")
	require.NoError(t, err)
	assert.Equal(t, []byte("c"), got)
}

// When no keys match the prefix the SCAN loop never executes; nil is returned.
func TestDeleteByPrefix_NoMatch_ReturnsNil(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "user:1", []byte("v"), longTTL))

	err := c.DeleteByPrefix(ctx, "token:")
	assert.NoError(t, err)
}

func TestDeleteByPrefix_EmptyCache_ReturnsNil(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.DeleteByPrefix(context.Background(), "any:")
	assert.NoError(t, err)
}

func TestDeleteByPrefix_ExactKeyMatch(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "exact", []byte("v"), longTTL))
	require.NoError(t, c.DeleteByPrefix(ctx, "exact"))

	_, err := c.Get(ctx, "exact")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDeleteByPrefix_SubstringNotPrefix_IsNotDeleted(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "abc", []byte("v1"), longTTL))
	require.NoError(t, c.Set(ctx, "xabc", []byte("v2"), longTTL))

	require.NoError(t, c.DeleteByPrefix(ctx, "ab"))

	_, err := c.Get(ctx, "abc")
	assert.ErrorIs(t, err, entities.ErrNotFound, "abc matches prefix 'ab'")

	got, err := c.Get(ctx, "xabc")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), got, "xabc does not start with 'ab'")
}

func TestDeleteByPrefix_EmptyPrefix_WipesCache(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "a", []byte("1"), longTTL))
	require.NoError(t, c.Set(ctx, "b", []byte("2"), longTTL))

	require.NoError(t, c.DeleteByPrefix(ctx, ""))

	_, err1 := c.Get(ctx, "a")
	_, err2 := c.Get(ctx, "b")
	assert.ErrorIs(t, err1, entities.ErrNotFound)
	assert.ErrorIs(t, err2, entities.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestGet_CancelledContext(t *testing.T) {
	c, _ := newTestCache(t)

	_, err := c.Get(cancelledCtx(), "k")
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded),
		"expected context error, got: %v", err)
}

func TestSet_CancelledContext(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.Set(cancelledCtx(), "k", []byte("v"), longTTL)
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded),
		"expected context error, got: %v", err)
}

func TestDelete_CancelledContext(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.Delete(cancelledCtx(), "k")
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded),
		"expected context error, got: %v", err)
}

// TestSet_CancelledContext_DoesNotWrite verifies that a cancelled-context Set
// does not persist the value in Redis.
func TestSet_CancelledContext_DoesNotWrite(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	_ = c.Set(cancelledCtx(), "k", []byte("v"), longTTL)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
