package job

import (
	"context"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"sync"
	"time"

	rlock "github.com/gotomicro/redis-lock"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.Logger
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client, l logger.Logger) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 根据数据量来，如果七天内贴子很多，可以适当调大
		timeout:   timeout,
		client:    client,
		key:       "rolock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "Ranking"
}

func (r *RankingJob) Run() error {
	// 本地锁，防止多个定时任务同时执行，防止并发
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 没拿到锁，极大概率别人持有了锁
			return nil
		}
		r.lock = lock
		// 保证这里一直拿到这个锁，使用协程自动续期
		go func() {
			// 本地锁，防止多个定时任务同时执行，防止并发
			r.localLock.Lock()
			defer r.localLock.Unlock()

			er := lock.AutoRefresh(r.timeout/2, time.Second)
			// 如果续期失败，说明锁被释放了
			if er != nil {
				// 如果失败了，就释放锁，争取下次能拿到
				r.lock = nil
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

// Close 关闭锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
