package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func newCache(t *testing.T) *MemoryCache {
	t.Helper()
	c, err := New()
	require.NoError(t, err)
	return c
}

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

const longTTL = time.Hour
const shortTTL = 5 * time.Millisecond
const shortSleep = 20 * time.Millisecond

// ---------------------------------------------------------------------------
// Set + Get happy path
// ---------------------------------------------------------------------------

func TestSetGet_StoresAndRetrievesValue(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("hello"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)
}

func TestSet_DoesNotBlock(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	// Set must return well within a reasonable deadline regardless of TTL value.
	deadline := time.After(100 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		_ = c.Set(ctx, "k", []byte("v"), longTTL)
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-deadline:
		t.Fatal("Set blocked longer than expected")
	}
}

// ---------------------------------------------------------------------------
// TTL edge cases
// ---------------------------------------------------------------------------

func TestGet_ExpiredEntry(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), shortTTL))
	time.Sleep(shortSleep)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestGet_ExpiredEntry_IsEvicted(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), shortTTL))
	time.Sleep(shortSleep)

	_, _ = c.Get(ctx, "k") // triggers lazy eviction

	c.mu.RLock()
	_, stillPresent := c.cache["k"]
	c.mu.RUnlock()
	assert.False(t, stillPresent, "expired entry should be removed from map after Get")
}

func TestGet_ZeroTTL_ImmediatelyExpired(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), 0))

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestGet_NegativeTTL_ImmediatelyExpired(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), -time.Second))

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestGet_LargeTTL(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), 24*365*time.Hour))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("v"), got)
}

func TestSet_OverwriteShortensTTL(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v1"), longTTL))
	require.NoError(t, c.Set(ctx, "k", []byte("v2"), shortTTL))
	time.Sleep(shortSleep)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestSet_OverwriteExtendsExpiredTTL(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v1"), shortTTL))
	time.Sleep(shortSleep)
	require.NoError(t, c.Set(ctx, "k", []byte("v2"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), got)
}

// ---------------------------------------------------------------------------
// Key / value edge cases
// ---------------------------------------------------------------------------

func TestSet_EmptyStringKey(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "", []byte("v"), longTTL))

	got, err := c.Get(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []byte("v"), got)
}

func TestSet_NilValue(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", nil, longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSet_EmptyByteSliceValue(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte{}, longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte{}, got)
}

func TestSet_OverwriteLastWriteWins(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("first"), longTTL))
	require.NoError(t, c.Set(ctx, "k", []byte("second"), longTTL))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), got)
}

// ---------------------------------------------------------------------------
// Get on missing / empty cache
// ---------------------------------------------------------------------------

func TestGet_MissingKey(t *testing.T) {
	c := newCache(t)
	_, err := c.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestGet_EmptyCache(t *testing.T) {
	c := newCache(t)
	_, err := c.Get(context.Background(), "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_ExistingKey(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), longTTL))
	require.NoError(t, c.Delete(ctx, "k"))

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDelete_MissingKey(t *testing.T) {
	c := newCache(t)
	err := c.Delete(context.Background(), "missing")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDelete_EmptyCache(t *testing.T) {
	c := newCache(t)
	err := c.Delete(context.Background(), "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDelete_AlreadyDeletedKey(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), longTTL))
	require.NoError(t, c.Delete(ctx, "k"))

	err := c.Delete(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// ---------------------------------------------------------------------------
// DeleteByPrefix
// ---------------------------------------------------------------------------

func TestDeleteByPrefix_MatchingKeys(t *testing.T) {
	c := newCache(t)
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

func TestDeleteByPrefix_NoMatch(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "user:1", []byte("v"), longTTL))

	err := c.DeleteByPrefix(ctx, "token:")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDeleteByPrefix_EmptyCache(t *testing.T) {
	c := newCache(t)
	err := c.DeleteByPrefix(context.Background(), "any:")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDeleteByPrefix_ExactKeyMatch(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "exact", []byte("v"), longTTL))
	require.NoError(t, c.DeleteByPrefix(ctx, "exact"))

	_, err := c.Get(ctx, "exact")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

func TestDeleteByPrefix_SubstringNotPrefix(t *testing.T) {
	// "ab" is a substring of "xabc" but not a prefix — must not be deleted.
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "abc", []byte("v1"), longTTL))
	require.NoError(t, c.Set(ctx, "xabc", []byte("v2"), longTTL))

	require.NoError(t, c.DeleteByPrefix(ctx, "ab"))

	_, err := c.Get(ctx, "abc")
	assert.ErrorIs(t, err, entities.ErrNotFound)

	got, err := c.Get(ctx, "xabc")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), got)
}

// TestDeleteByPrefix_EmptyPrefix documents that an empty prefix matches every
// key, effectively wiping the entire cache.
func TestDeleteByPrefix_EmptyPrefix_WipesCache(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "a", []byte("1"), longTTL))
	require.NoError(t, c.Set(ctx, "b", []byte("2"), longTTL))

	require.NoError(t, c.DeleteByPrefix(ctx, ""))

	_, err1 := c.Get(ctx, "a")
	_, err2 := c.Get(ctx, "b")
	assert.ErrorIs(t, err1, entities.ErrNotFound)
	assert.ErrorIs(t, err2, entities.ErrNotFound)
}

func TestDeleteByPrefix_IncludesExpiredEntries(t *testing.T) {
	// Expired entries are still in the map until lazily evicted; DeleteByPrefix
	// should delete them and not return ErrNotFound.
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "user:expired", []byte("v"), shortTTL))
	time.Sleep(shortSleep)

	err := c.DeleteByPrefix(ctx, "user:")
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestGet_CancelledContext(t *testing.T) {
	c := newCache(t)
	_, err := c.Get(cancelledCtx(), "k")
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func TestSet_CancelledContext(t *testing.T) {
	c := newCache(t)
	err := c.Set(cancelledCtx(), "k", []byte("v"), longTTL)
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func TestDelete_CancelledContext(t *testing.T) {
	c := newCache(t)
	err := c.Delete(cancelledCtx(), "k")
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func TestDeleteByPrefix_CancelledContext(t *testing.T) {
	c := newCache(t)
	err := c.DeleteByPrefix(cancelledCtx(), "k")
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func TestSet_CancelledContext_DoesNotWrite(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()

	_ = c.Set(cancelledCtx(), "k", []byte("v"), longTTL)

	_, err := c.Get(ctx, "k")
	assert.ErrorIs(t, err, entities.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Concurrency (run with -race)
// ---------------------------------------------------------------------------

func TestConcurrent_SetGet_DistinctKeys(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	const n = 100

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i%26))
			_ = c.Set(ctx, key, []byte{byte(i)}, longTTL)
			_, _ = c.Get(ctx, key)
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_SetGet_SameKey(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	const n = 100

	var wg sync.WaitGroup
	for range n {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = c.Set(ctx, "k", []byte("v"), longTTL)
		}()
		go func() {
			defer wg.Done()
			_, _ = c.Get(ctx, "k")
		}()
	}
	wg.Wait()
}

func TestConcurrent_DoubleEviction_ExpiredKey(t *testing.T) {
	// Two goroutines racing to evict the same expired key must not panic or deadlock.
	c := newCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("v"), shortTTL))
	time.Sleep(shortSleep)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(ctx, "k")
		}()
	}
	wg.Wait()
}

func TestConcurrent_DeleteByPrefix_And_Set(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	const n = 50

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			_ = c.Set(ctx, "pfx:key", []byte{byte(i)}, longTTL)
		}(i)
		go func() {
			defer wg.Done()
			_ = c.DeleteByPrefix(ctx, "pfx:")
		}()
	}
	wg.Wait()
}
