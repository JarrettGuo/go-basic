package ioc

import (
	"go-basic/webook/internal/job"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankingService) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30)
}

func InitJob(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		l.Error("添加任务失败", logger.Error(err))
	}
	return res
}
