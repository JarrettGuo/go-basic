package article

import (
	"context"
	"go-basic/webook/internal/domain"
	dao "go-basic/webook/internal/repository/dao/article"

	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
}

type CacheArticleRepository struct {
	dao dao.ArticleDAO

	reader dao.ReaderDAO
	author dao.AuthorDAO
	db     *gorm.DB
}

func NewArticleRepository(dao dao.ArticleDAO, reader dao.ReaderDAO, author dao.AuthorDAO) ArticleRepository {
	return &CacheArticleRepository{
		dao:    dao,
		reader: reader,
		author: author,
	}
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, c.domainToEntity(ctx, art))
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.domainToEntity(ctx, art))
}

// SyncV2 同步文章，使用事务
func (c *CacheArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 事务结束或失败时，回滚
	defer tx.Rollback()
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)
	var (
		id  = art.Id
		err error
	)
	artEntity := c.domainToEntity(ctx, art)
	if id > 0 {
		err = author.UpdateById(ctx, artEntity)
	} else {
		id, err = author.Insert(ctx, artEntity)
	}
	if err != nil {
		return 0, err
	}
	// 同步线上库，使用不同的表
	err = reader.UpsertV2(ctx, dao.PublishArticle{
		Article: artEntity})
	// 执行成功，提交
	tx.Commit()
	return id, err
}

func (c *CacheArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artEntity := c.domainToEntity(ctx, art)
	if id > 0 {
		err = c.author.UpdateById(ctx, artEntity)
	} else {
		id, err = c.author.Insert(ctx, artEntity)
	}
	if err != nil {
		return 0, err
	}
	// 同步线上库
	err = c.reader.Upsert(ctx, artEntity)
	return id, err
}

func (c *CacheArticleRepository) domainToEntity(ctx context.Context, art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}

func (c *CacheArticleRepository) entityToDomain(ctx context.Context, art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
	}
}
