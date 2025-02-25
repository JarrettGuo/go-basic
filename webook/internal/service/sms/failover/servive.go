package failover

import (
	"context"
	"errors"
	"go-basic/webook/internal/service/sms"
	"sync/atomic"
)

type FailoverSMSService struct {
	svcs []sms.Service

	idx uint64
}

func NewFailoverSMSService(svcs ...sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 取下一个节点作为起始节点
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(ctx, tpl, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		default:
			// 输出日志
		}
	}
	return errors.New("所有服务都失败了")
}
