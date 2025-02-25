package failover

import (
	"context"
	"go-basic/webook/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	// 你的服务列表
	svcs []sms.Service
	idx  int32
	// 连续超时次数
	cnt int32
	// 阈值，连续超时多少次后，切换到下一个服务
	threshold int32
}

func NewTimeoutFailoverSMSService() sms.Service {

	return &TimeoutFailoverSMSService{}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		// 切换到下一个服务，新的下标
		newIdx := (idx + 1) % int32(len(t.svcs))
		// 尝试切换，如果成功，就把计数器清零
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 我成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现了并发问题，别人换成功了
		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		// 超时了，计数器加一
		atomic.AddInt32(&t.cnt, 1)
	case nil:
		// 成功了，计数器清零
		atomic.StoreInt32(&t.cnt, 0)
	default:
		// 其他错误，不做处理
		return err
	}
	return nil
}
