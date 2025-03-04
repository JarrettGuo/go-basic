package article

import (
	"context"
	"go-basic/webook/internal/domain"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, art domain.Article) error
}
