package ioc

import (
	"go-basic/webook/internal/job"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankingService, l logger.Logger, rlockClient *rlock.Client) (*job.RankingJob, func()) {
	j := job.NewRankingJob(svc, time.Second*30, rlockClient, l)
	return j, func() {
		j.Close()
	}
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
