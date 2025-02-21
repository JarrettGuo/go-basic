package service

import (
	"context"
	"fmt"
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/service/sms"

	"golang.org/x/exp/rand"
)

const codeTplId = "123456"

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, code string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成验证码
	code := svc.generateCode()
	// 存储验证码
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送验证码
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	return err
}

func (svc *codeService) Verify(ctx context.Context, biz string, phone string, code string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, code)
}

func (svc *codeService) generateCode() string {
	// 生成验证码
	num := rand.Intn(1000000)
	// 不够6位数前面补0
	return fmt.Sprintf("%06d", num)
}
