package service

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
)

type RankingService interface {
	TopN(ctx context.Context) error
	topN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	repo      repository.RankingRepository
	artSvc    ArticleService
	intrSvc   InteractiveService
	batchSize int
	n         int
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

func (s *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := s.topN(ctx)
	if err != nil {
		return err
	}
	// 存起来
	return s.repo.ReplaceTopN(ctx, arts)
}

func (s *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	// 先获取一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	topN := queue.NewConcurrentPriorityQueue[Score](s.n, func(i, j Score) int {
		if i.score > j.score {
			return 1
		} else if i.score < j.score {
			return -1
		}
		return 0
	})
	for {
		// 先获取一批数据
		arts, err := s.artSvc.ListPub(ctx, offset, s.batchSize, now)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		// 找到对应的点赞数据
		intrs, err := s.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 合并计算score
		// 排序
		for _, art := range arts {
			intr, ok := intrs[art.Id]
			if !ok {
				// 没有点赞数据, 跳过
				continue
			}
			score := s.scoreFunc(art.Utime, intr.LikeCnt)
			// 考虑这个score是否在前一百名
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Peek()
				if score > val.score {
					topN.Dequeue()
					topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				}
			}
		}
		// 一批已经处理完了，看看是否还有下一批，当数据不足或者数据太旧时，停止
		if len(arts) < s.batchSize || now.Sub(arts[0].Utime) > 7*24*time.Hour {
			break
		}
		// 更新offset
		offset = offset + len(arts)
	}
	// 得出结果
	res := make([]domain.Article, s.n)
	for i := s.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}
