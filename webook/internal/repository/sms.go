package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/dao"
	"strings"
	"time"
)

type SMSRepository interface {
	Store(ctx context.Context, tpl string, args []string, numbers []string) error
}

type SMSAysncReqRepository struct {
	dao dao.SMSAysncReqDAO
}

func NewSMSAysncReqRepository(dao dao.SMSAysncReqDAO) SMSRepository {
	return &SMSAysncReqRepository{
		dao: dao,
	}
}

func (repo *SMSAysncReqRepository) Store(ctx context.Context, tpl string, args []string, numbers []string) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(
		domain.SMS{
			Biz:     tpl,
			Args:    args,
			Numbers: numbers,
			Status:  0,
			Ctime:   time.Now(),
		},
	))
}

func (repo *SMSAysncReqRepository) Find(ctx context.Context) ([]domain.SMS, error) {
	reqs, err := repo.dao.FindByStatus(ctx, 0)
	if err != nil {
		return nil, err
	}
	res := make([]domain.SMS, 0, len(reqs))
	for _, req := range reqs {
		res = append(res, repo.entityToDomain(req))
	}
	return res, nil
}

func (repo *SMSAysncReqRepository) entityToDomain(req dao.SMSAysncReq) domain.SMS {
	return domain.SMS{
		Id:       req.Id,
		Biz:      req.Biz,
		Args:     []string{req.Args},
		Numbers:  []string{req.Numbers},
		Status:   req.Status,
		RetryCnt: req.RetryCnt,
		Ctime:    time.UnixMilli(req.Ctime),
	}
}

func (repo *SMSAysncReqRepository) domainToEntity(req domain.SMS) dao.SMSAysncReq {
	return dao.SMSAysncReq{
		Id:       req.Id,
		Args:     strings.Join(req.Args, ","),
		Numbers:  strings.Join(req.Numbers, ","),
		Status:   req.Status,
		RetryCnt: req.RetryCnt,
		Ctime:    req.Ctime.UnixMilli(),
	}
}
