package article

import (
	"context"
	"go-basic/webook/internal/domain"
	userRepo "go-basic/webook/internal/repository"
	"go-basic/webook/internal/repository/cache"
	dao "go-basic/webook/internal/repository/dao/article"
	"go-basic/webook/pkg/logger"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	// SyncV2(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, offset, limit int, start time.Time) ([]domain.Article, error)
}

type CacheArticleRepository struct {
	// dao dao.ArticleDAO
	dao      dao.ArticleDAO
	userRepo userRepo.UserRepository

	reader dao.ReaderDAO
	author dao.AuthorDAO
	db     *gorm.DB

	cache cache.ArticleCache
	l     logger.Logger
}

func NewArticleRepository(dao dao.ArticleDAO, reader dao.ReaderDAO, author dao.AuthorDAO, cache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CacheArticleRepository{
		dao:    dao,
		reader: reader,
		author: author,
		cache:  cache,
		l:      l,
	}
}

func (c *CacheArticleRepository) ListPub(ctx context.Context, offset, limit int, start time.Time) ([]domain.Article, error) {
	res, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.Article) domain.Article {
		return c.entityToDomain(ctx, src)
	}), nil
}

func (c *CacheArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果内容放在oss上，让前端直接访问oss
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 组装 user，适合单体架构
	user, err := c.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (c *CacheArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.entityToDomain(ctx, data), nil
}

func (c *CacheArticleRepository) List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error) {
	// 进行缓存
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return c.entityToDomain(ctx, src)
	})
	// 设置回写缓存，可以同步或异步
	go func() {
		err := c.cache.SetFirstPage(ctx, uid, data)
		c.l.Error("回写缓存失败", logger.Error(err))
		c.preCache(ctx, data)
	}()
	return data, nil
}

func (c *CacheArticleRepository) SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, authorId, status.ToUint8())
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 删除缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.Insert(ctx, c.domainToEntity(ctx, art))
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 删除缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.UpdateById(ctx, c.domainToEntity(ctx, art))
}

// SyncV2 同步文章，使用事务
// func (c *CacheArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
// 	// 开启一个事务
// 	tx := c.db.WithContext(ctx).Begin()
// 	if tx.Error != nil {
// 		return 0, tx.Error
// 	}
// 	// 事务结束或失败时，回滚
// 	defer tx.Rollback()
// 	author := dao.NewAuthorDAO(tx)
// 	reader := dao.NewReaderDAO(tx)
// 	var (
// 		id  = art.Id
// 		err error
// 	)
// 	artEntity := c.domainToEntity(ctx, art)
// 	if id > 0 {
// 		err = author.UpdateById(ctx, artEntity)
// 	} else {
// 		id, err = author.Insert(ctx, artEntity)
// 	}
// 	if err != nil {
// 		return 0, err
// 	}
// 	// 同步线上库，使用不同的表
// 	err = reader.Upsert(ctx, dao.PublishedArticle{
// 		Article: artEntity})
// 	// 执行成功，提交
// 	tx.Commit()
// 	return id, err
// }

func (c *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.domainToEntity(ctx, art))
	if err == nil {
		c.cache.DelFirstPage(ctx, art.Author.Id)
		c.cache.SetPub(ctx, art)
	}
	return id, err
}

func (c *CacheArticleRepository) domainToEntity(ctx context.Context, art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
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
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}

// 预缓存
func (c *CacheArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("预缓存失败", logger.Error(err))
		}
	}
}
