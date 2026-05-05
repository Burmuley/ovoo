package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/redis/go-redis/v9"
)

// RedisCache is a cache.Cache implementation backed by a Redis server.
// It wraps the go-redis client and translates Redis-specific errors into
// the domain error types used by the rest of the application.
type RedisCache struct {
	client *redis.Client
}

// New creates a RedisCache from the supplied APICacheConfig.
// It reads the Redis sub-section of the config (address, credentials, DB index,
// and RESP protocol version) and constructs a go-redis client. The client
// connection is lazy — no network dial occurs until the first command is issued.
// Protocol defaults to RESP3 (version 3) when the config value is zero.
func New(config config.APICacheConfig) (*RedisCache, error) {
	// parse configuration
	redisCfg := config.Config.Redis
	opts := redis.Options{}
	if redisCfg.Addr != nil {
		opts.Addr = *redisCfg.Addr
	}

	if redisCfg.Username != nil {
		opts.Username = *redisCfg.Username
	}

	if redisCfg.Password != nil {
		opts.Password = *redisCfg.Password
	}

	if redisCfg.Protocol != 0 {
		opts.Protocol = redisCfg.Protocol
	} else {
		opts.Protocol = 3 // default protocol version
	}

	opts.DB = redisCfg.DB

	// init client
	cli := redis.NewClient(&opts)
	cache := &RedisCache{client: cli}
	return cache, nil
}

// Get returns the value stored under key. It returns entities.ErrNotFound when
// the key does not exist or has expired, and propagates context errors unchanged.
// Any other Redis error is mapped to entities.ErrDatabase.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, wrapRedisErr(err)
	}

	return []byte(res), nil
}

// Set stores value under key with the given TTL. A zero or negative TTL causes
// Redis to expire the key immediately. Errors are translated via wrapRedisErr.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	_, err := c.client.Set(ctx, key, value, ttl).Result()
	return wrapRedisErr(err)
}

// Delete removes key from the cache. Because Redis DEL is idempotent it returns
// nil even when the key does not exist — no entities.ErrNotFound is raised.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, key).Result()
	return wrapRedisErr(err)
}

// DeleteByPrefix removes every key whose name starts with prefix using a SCAN
// + DEL loop. It returns nil when no keys match (the loop body never executes).
// Note: context cancellation that occurs before the SCAN completes is not
// propagated — only errors from individual DEL commands are returned.
func (c *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	iter := c.client.Scan(ctx, 0, prefix+"*", 0).Iterator()

	for iter.Next(ctx) {
		err := c.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			return wrapRedisErr(err)
		}
	}

	return nil
}

// wrapRedisErr maps Redis-specific errors to domain errors:
//   - nil → nil
//   - context.Canceled / context.DeadlineExceeded → returned as-is
//   - redis.Nil (key not found) → entities.ErrNotFound
//   - anything else → entities.ErrDatabase
func wrapRedisErr(err error) error {
	if err == nil {
		return err
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	if errors.Is(err, redis.Nil) {
		return entities.ErrNotFound
	}

	return fmt.Errorf("%w: %w", entities.ErrDatabase, err)
}
