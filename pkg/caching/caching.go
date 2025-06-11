package caching

import "context"

type Caching interface {
	IncrBy(ctx context.Context, key string, value int64) error
	DecrBy(ctx context.Context, key string, value int64) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
}
