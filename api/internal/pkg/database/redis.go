package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/logger"
)

// RedisDB wraps a Redis client
type RedisDB struct {
	Client *redis.Client
}

// NewRedis creates a new Redis client
func NewRedis(ctx context.Context, cfg config.RedisConfig) (*RedisDB, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        100,
		MinIdleConns:    10,
		PoolTimeout:     4 * time.Second,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("connected to Redis",
		zap.String("addr", addr),
		zap.Int("db", cfg.DB),
	)

	return &RedisDB{Client: client}, nil
}

// Close closes the Redis connection
func (db *RedisDB) Close() error {
	if db.Client != nil {
		return db.Client.Close()
	}
	return nil
}

// Get gets a value by key
func (db *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return db.Client.Get(ctx, key).Result()
}

// Set sets a value with optional expiration
func (db *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return db.Client.Set(ctx, key, value, expiration).Err()
}

// SetNX sets a value only if it doesn't exist
func (db *RedisDB) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return db.Client.SetNX(ctx, key, value, expiration).Result()
}

// Del deletes one or more keys
func (db *RedisDB) Del(ctx context.Context, keys ...string) error {
	return db.Client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (db *RedisDB) Exists(ctx context.Context, keys ...string) (int64, error) {
	return db.Client.Exists(ctx, keys...).Result()
}

// Incr increments a key
func (db *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return db.Client.Incr(ctx, key).Result()
}

// IncrBy increments a key by a value
func (db *RedisDB) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return db.Client.IncrBy(ctx, key, value).Result()
}

// Expire sets expiration on a key
func (db *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return db.Client.Expire(ctx, key, expiration).Err()
}

// TTL gets the remaining time to live of a key
func (db *RedisDB) TTL(ctx context.Context, key string) (time.Duration, error) {
	return db.Client.TTL(ctx, key).Result()
}

// HGet gets a hash field
func (db *RedisDB) HGet(ctx context.Context, key, field string) (string, error) {
	return db.Client.HGet(ctx, key, field).Result()
}

// HSet sets hash fields
func (db *RedisDB) HSet(ctx context.Context, key string, values ...interface{}) error {
	return db.Client.HSet(ctx, key, values...).Err()
}

// HGetAll gets all hash fields
func (db *RedisDB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return db.Client.HGetAll(ctx, key).Result()
}

// HIncrBy increments a hash field by value
func (db *RedisDB) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return db.Client.HIncrBy(ctx, key, field, incr).Result()
}

// Publish publishes a message to a channel
func (db *RedisDB) Publish(ctx context.Context, channel string, message interface{}) error {
	return db.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (db *RedisDB) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return db.Client.Subscribe(ctx, channels...)
}

// SAdd adds members to a set
func (db *RedisDB) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return db.Client.SAdd(ctx, key, members...).Err()
}

// SMembers gets all set members
func (db *RedisDB) SMembers(ctx context.Context, key string) ([]string, error) {
	return db.Client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member is in a set
func (db *RedisDB) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return db.Client.SIsMember(ctx, key, member).Result()
}

// ZAdd adds members to a sorted set
func (db *RedisDB) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return db.Client.ZAdd(ctx, key, members...).Err()
}

// ZRangeWithScores gets members with scores from sorted set
func (db *RedisDB) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return db.Client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// Pipeline returns a pipeline for batch operations
func (db *RedisDB) Pipeline() redis.Pipeliner {
	return db.Client.Pipeline()
}

// RateLimit implements a simple rate limiter using Redis
func (db *RedisDB) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	pipe := db.Client.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := incr.Val()
	if count > limit {
		return false, limit - count, nil
	}

	return true, limit - count, nil
}

// Cache implements a simple cache with TTL
type Cache struct {
	redis *RedisDB
	ttl   time.Duration
}

// NewCache creates a new cache
func NewCache(redis *RedisDB, ttl time.Duration) *Cache {
	return &Cache{
		redis: redis,
		ttl:   ttl,
	}
}

// Get gets a cached value
func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	val, err := c.redis.Get(ctx, key)
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return val, true
}

// Set sets a cached value
func (c *Cache) Set(ctx context.Context, key, value string) error {
	return c.redis.Set(ctx, key, value, c.ttl)
}

// Delete deletes a cached value
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key)
}
