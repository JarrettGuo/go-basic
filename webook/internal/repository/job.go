package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/dao"
	"time"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64, version int) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type PreemptCronJobRepository struct {
	dao dao.JobDAO
}

func NewPreemptCronJobRepository(dao dao.JobDAO) *PreemptCronJobRepository {
	return &PreemptCronJobRepository{
		dao: dao,
	}
}

func (p *PreemptCronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, next)
}

func (p *PreemptCronJobRepository) Release(ctx context.Context, id int64, version int) error {
	return p.dao.Release(ctx, id, version)
}

func (p *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Id:       j.Id,
		Cfg:      j.Cfg,
		Version:  j.Version,
		Name:     j.Name,
		Executor: j.Executor,
	}, nil
}
