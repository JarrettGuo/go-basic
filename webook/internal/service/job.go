package service

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository"
	"go-basic/webook/pkg/logger"
	"time"
)

type JobService interface {
	// 抢占任务
	Preempt(ctx context.Context) (domain.Job, error)
	refresh(id int64)
	ResetNextTime(ctx context.Context, job domain.Job) error
}

type CronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.Logger
}

func (p *CronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.repo.Preempt(ctx)

	// 续约
	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for range ticker.C {
			p.refresh(j.Id)
		}
	}()

	// 抢占之后，考虑释放资源
	version := j.Version
	j.CancelFunc = func() error {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return p.repo.Release(ctx, j.Id, version)
	}
	return j, err
}

func (p *CronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 续约，更新一下时间就可以
	err := p.repo.UpdateUtime(ctx, id)
	if err != nil {
		p.l.Error("续约失败", logger.Error(err), logger.Int64("job_id", id))
	}
}

func (p *CronJobService) ResetNextTime(ctx context.Context, job domain.Job) error {
	next := job.NextTime()
	if next.IsZero() {
		return p.repo.Stop(ctx, job.Id)
	}
	return p.repo.UpdateNextTime(ctx, job.Id, next)
}
