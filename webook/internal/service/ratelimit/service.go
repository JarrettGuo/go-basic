package ratelimit

import (
	"context"
	"fmt"
	"go-basic/webook/internal/service/sms"
	"go-basic/webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("短信服务限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s RatelimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流失败: %w", err)
	}
	if limited {
		return errLimited
	}
	err = s.svc.Send(ctx, tplId, args, numbers...)
	return err
}
