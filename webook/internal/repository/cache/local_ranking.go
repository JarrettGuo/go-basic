package cache

import (
	"context"
	"errors"
	"go-basic/webook/internal/domain"
	"time"

	"github.com/ecodeclub/ekit/syncx/atomicx"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewRankingLocalCache() *RankingLocalCache {
	return &RankingLocalCache{
		topN: atomicx.NewValue[[]domain.Article](),
		ddl:  atomicx.NewValueOf(time.Now()),
		// 本地缓存过期时间非常长，或永不过期
		expiration: time.Hour * 24,
	}
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	r.topN.Store(arts)
	ddl := time.Now().Add(r.expiration)
	r.ddl.Store(ddl)
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	arts := r.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存不存在或已过期")
	}
	return r.topN.Load(), nil
}

func (r *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	return r.topN.Load(), nil
}
