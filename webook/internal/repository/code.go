package repository

import (
	"context"
	"go-basic/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany = cache.ErrCodeSendTooMany
)

type CodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: cache,
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}
