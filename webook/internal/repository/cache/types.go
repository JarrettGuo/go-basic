package cache

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit"
)

type Cache interface {
	Set(ctx context.Context, key string, val any, exp time.Duration) error
	Get(ctx context.Context, key string) (ekit.AnyValue, error)
}

type LocalCache struct {
}

type RedisCache struct{}

type DoubleCache struct {
	local Cache
	redis Cache
}

func (d *DoubleCache) set(ctx context.Context, key string, val any, exp time.Duration) error {
	panic("implement me")
}

func (d *DoubleCache) Get(ctx context.Context, key string) (ekit.AnyValue, error) {
	panic("implement me")
}
