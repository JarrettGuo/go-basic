package article

import (
	"context"

	"gorm.io/gorm"
)

type ReaderDAO interface {
	// Upsert(ctx context.Context, art Article) error
	Upsert(ctx context.Context, art PublishedArticle) error
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

type PublishedArticle struct {
	Article
}
