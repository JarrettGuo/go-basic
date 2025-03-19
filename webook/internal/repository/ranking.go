package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
}

type CacheRankingRepository struct {
	c cache.RankingCache
}

func NewRankingRepository(c cache.RankingCache) RankingRepository {
	return &CacheRankingRepository{
		c: c,
	}
}

func (c *CacheRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.c.Set(ctx, arts)
}
