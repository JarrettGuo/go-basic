package async

import (
	"context"
	"errors"
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/service/sms"
	"net"
	"strings"
	"syscall"
)

type SMSService struct {
	svc  sms.Service
	repo repository.SMSAysncReqRepository
}

func NewSMSService(svc sms.Service) *SMSService {
	return &SMSService{
		svc: svc,
	}
}

func (s *SMSService) StartAysnc(ctx context.Context) {
	go func() {
		lists, err := s.repo.Find(ctx)
		if err != nil {
			return
		}
		for _, req := range lists {
			_ = s.svc.Send(ctx, req.Biz, req.Args, req.Numbers...)
		}
	}()
}

func (s *SMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 正常路径
	err := s.svc.Send(ctx, tpl, args, numbers...)
	// 异常路径
	if err != nil {
		if s.shouldRetryAsync(err) {
			return s.repo.Store(ctx, tpl, args, numbers)
		}
		return err
	}
	return nil
}

func (s *SMSService) shouldRetryAsync(err error) bool {
	if s.isCommonTemporaryError(err) {
		return true
	}

	errMsg := err.Error()
	// 服务端错误
	if strings.Contains(strings.ToLower(errMsg), "service unavailable") ||
		strings.Contains(strings.ToLower(errMsg), "server busy") ||
		strings.Contains(strings.ToLower(errMsg), "try again later") {
		return true
	}

	// 限流错误
	if strings.Contains(strings.ToLower(errMsg), "rate limit") ||
		strings.Contains(strings.ToLower(errMsg), "too many requests") ||
		strings.Contains(strings.ToLower(errMsg), "throttl") ||
		strings.Contains(strings.ToLower(errMsg), "limit exceed") {
		return true
	}

	// 网络错误
	if strings.Contains(strings.ToLower(errMsg), "connection") ||
		strings.Contains(strings.ToLower(errMsg), "network") ||
		strings.Contains(strings.ToLower(errMsg), "timeout") {
		return true
	}

	// 不需要重试的错误类型
	if strings.Contains(strings.ToLower(errMsg), "invalid parameter") ||
		strings.Contains(strings.ToLower(errMsg), "unauthorized") ||
		strings.Contains(strings.ToLower(errMsg), "authentication failed") ||
		strings.Contains(strings.ToLower(errMsg), "sensitive word") ||
		strings.Contains(strings.ToLower(errMsg), "black list") ||
		strings.Contains(strings.ToLower(errMsg), "daily limit") ||
		strings.Contains(strings.ToLower(errMsg), "content illegal") {
		return false
	}

	return true
}

func (s *SMSService) isCommonTemporaryError(err error) bool {
	// 上下文超时
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// 网络错误
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	// 连接错误
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	return false
}
