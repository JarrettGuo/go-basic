package ioc

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/job"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"time"
)

func InitScheduler(l logger.Logger, svc service.JobService, local *job.LocalFuncExecter) *job.Scheduler {
	res := job.NewScheduler(svc, l)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor(svc service.RankingService) *job.LocalFuncExecter {
	res := job.NewLocalFuncExecter()
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}
