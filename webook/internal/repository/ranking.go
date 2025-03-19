package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CacheRankingRepository struct {
	redis cache.RankingCache
	local cache.RankingLocalCache
}

func NewRankingRepository(redis cache.RankingCache, local cache.RankingLocalCache) RankingRepository {
	return &CacheRankingRepository{
		redis: redis,
		local: local,
	}
}

func (c *CacheRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.local.Set(ctx, arts)
	return c.redis.Set(ctx, arts)
}

func (c *CacheRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := c.local.Get(ctx)
	if err == nil {
		return data, nil
	}
	data, err = c.redis.Get(ctx)
	if err == nil {
		c.local.Set(ctx, data)
	} else {
		return c.local.ForceGet(ctx)
	}
	return data, err
}
