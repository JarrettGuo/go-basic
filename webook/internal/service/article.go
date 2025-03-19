package service

import (
	"context"
	events "go-basic/webook/events/article"
	"go-basic/webook/internal/domain"
	repository "go-basic/webook/internal/repository/article"
	"go-basic/webook/pkg/logger"
	"time"
)

//go:generate mockgen -source=article.go -package=svcmocks -destination=mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, offset, limit int, start time.Time) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	author   repository.ArticleAuthorRepository
	reader   repository.ArticleReaderRepository
	l        logger.Logger
	producer events.Producer
	ch       chan readInfo
}

type readInfo struct {
	uid int64
	aid int64
}

func NewArticleService(repo repository.ArticleRepository, l logger.Logger, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository, reader repository.ArticleReaderRepository, l logger.Logger) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func NewArticleServiceV2(repo repository.ArticleRepository, l logger.Logger, producer events.Producer) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		for {
			uids := make([]int64, 0, 10)
			aids := make([]int64, 0, 10)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			for i := 0; i < 10; i++ {
				select {
				case info, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					uids = append(uids, info.uid)
					aids = append(aids, info.aid)
				case <-ctx.Done():
					break
				}
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			producer.ProduceReadEventV1(ctx, events.ReadEventV1{
				Uids: uids,
				Aids: aids,
			})
			cancel()
		}
	}()
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
		ch:       make(chan readInfo, 10),
	}
}

func (a *articleService) ListPub(ctx context.Context, offset, limit int, start time.Time) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, offset, limit, start)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := a.repo.GetPublishedById(ctx, id)
	if err != nil {
		go func() {
			er := a.producer.ProduceReadEvent(ctx, events.ReadEvent{
				// 即使消费者需要使用art中数据，让他去查询
				Uid: uid,
				Aid: id,
			})
			if er != nil {
				a.l.Error("发送阅读事件失败", logger.Error(er), logger.Int64("art_id", id), logger.Int64("uid", uid))
			}
		}()

		// 改批量
		go func() {
			a.ch <- readInfo{
				uid: uid,
				aid: id,
			}
		}()
	}
	return art, err
}

func (a *articleService) List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}

	id, err := a.repo.Create(ctx, art)
	return id, err
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	// 重试
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("部分失败，保存到线上库失败", logger.Int64("art_id", id), logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，重试彻底失败", logger.Int64("art_id", id), logger.Error(err))
	}
	return id, err
}
