package service

import (
	"context"
	"go-basic/webook/internal/domain"
	repository "go-basic/webook/internal/repository/article"
	"go-basic/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo   repository.ArticleRepository
	author repository.ArticleAuthorRepository
	reader repository.ArticleReaderRepository
	l      logger.Logger
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository, reader repository.ArticleReaderRepository, l logger.Logger) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.SyncV1(ctx, art)
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
