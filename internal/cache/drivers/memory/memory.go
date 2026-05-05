package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
)

// memValue holds a cached byte slice together with its absolute expiry instant.
type memValue struct {
	value []byte
	ttl   time.Time
}

// MemoryCache is an in-process, goroutine-safe cache.Cache implementation.
// Entries are stored in a plain Go map and expire lazily on the next Get.
// It carries no background goroutine, so unused entries are only evicted
// when they are accessed after their TTL has elapsed.
type MemoryCache struct {
	mu    sync.RWMutex
	cache map[string]memValue
}

// New allocates and returns an empty MemoryCache ready for use.
func New() (*MemoryCache, error) {
	return &MemoryCache{cache: make(map[string]memValue)}, nil
}

// Get returns the value stored under key.
// It returns entities.ErrNotFound when the key is absent or its TTL has elapsed;
// in the latter case the entry is also deleted (lazy eviction).
// Returns ctx.Err() immediately if the context is already done.
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	item, ok := c.cache[key]
	c.mu.RUnlock()

	if !ok {
		return nil, entities.ErrNotFound
	}

	if time.Now().After(item.ttl) {
		_ = c.Delete(ctx, key)
		return nil, entities.ErrNotFound
	}

	return item.value, nil
}

// Set stores value under key, replacing any previous entry.
// The entry expires at now+ttl; a zero or negative ttl makes it immediately expired.
// Returns ctx.Err() immediately if the context is already done.
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	c.cache[key] = memValue{
		value: value,
		ttl:   time.Now().Add(ttl),
	}
	c.mu.Unlock()

	return nil
}

// Delete removes key from the cache.
// Returns entities.ErrNotFound if the key does not exist (including already-expired entries
// that have already been evicted). Returns ctx.Err() if the context is done.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; !ok {
		return entities.ErrNotFound
	}

	delete(c.cache, key)
	return nil
}

// DeleteByPrefix removes all keys that start with prefix in a single write-locked pass.
// Returns entities.ErrNotFound when no matching keys are found (including an empty cache).
// An empty prefix matches every key, effectively wiping the cache.
// Returns ctx.Err() if the context is done.
func (c *MemoryCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delcnt := 0
	for key := range c.cache {
		if strings.HasPrefix(key, prefix) {
			delete(c.cache, key)
			delcnt++
		}
	}

	if delcnt == 0 {
		return entities.ErrNotFound
	}

	return nil
}
