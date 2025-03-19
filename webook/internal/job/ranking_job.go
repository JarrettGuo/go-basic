package job

import (
	"context"
	"go-basic/webook/internal/service"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
}

func NewRankingJob(svc service.RankingService, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 根据数据量来，如果七天内贴子很多，可以适当调大
		timeout: timeout,
	}
}

func (r *RankingJob) Name() string {
	return "Ranking"
}

func (r *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}
