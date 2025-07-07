package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements both Storage and RateLimitStorage interfaces
type RedisStorage struct {
	client *redis.Client
}

// Ensure RedisStorage implements the interfaces
var _ Storage = (*RedisStorage)(nil)
var _ RateLimitStorage = (*RedisStorage)(nil)

func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{client: client}, nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", err
	}
	return result, err
}

func (r *RedisStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

func (r *RedisStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.IncrBy(ctx, key, by)
	pipe.Expire(ctx, key, 24*time.Hour) // Default TTL if not set

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Rate Limit specific methods
func (r *RedisStorage) GetTokens(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

func (r *RedisStorage) SetTokens(ctx context.Context, key string, tokens int64, ttl time.Duration) error {
	return r.client.Set(ctx, key, tokens, ttl).Err()
}

func (r *RedisStorage) DecrementTokens(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Decr(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

func (r *RedisStorage) IncrementTokens(ctx context.Context, key string, by int64) (int64, error) {
	result, err := r.client.IncrBy(ctx, key, by).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}
