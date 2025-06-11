package caching

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	return &RedisClient{client: rdb}
}

func (c *RedisClient) IncrBy(ctx context.Context, key string, value int64) error {
	return c.client.IncrBy(ctx, key, value).Err()
}

func (c *RedisClient) DecrBy(ctx context.Context, key string, value int64) error {
	return c.client.DecrBy(ctx, key, value).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, key, value, 0).Err()
}
