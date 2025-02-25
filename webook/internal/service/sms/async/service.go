package async

import (
	"context"
	"go-basic/webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	// repo repository.SMSAysncReqRepository
}

func NewSMSService(svc sms.Service) *SMSService {
	return &SMSService{
		svc: svc,
	}
}

func (s *SMSService) StartAysnc() {
	go func() {
		//s.repo.Find()   没发出去的请求
		// 遍历发送
	}()
}

func (s *SMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 正常路径

	// 异常路径
	err := s.svc.Send(ctx, tpl, args, numbers...)
	if err != nil {
		// 判断是否崩溃

		// if 崩溃 { s.repo.store()} 存入数据库
	}
	return nil
}
