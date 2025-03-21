package job

import (
	"context"
	"fmt"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"time"

	"golang.org/x/sync/semaphore"
)

type Executor interface {
	// 任务名称
	Name() string
	// 具体执行任务
	Exec(ctx context.Context, j domain.Job) error
	// 注册任务执行函数
	RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error)
}

type LocalFuncExecter struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecter() *LocalFuncExecter {
	return &LocalFuncExecter{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

func (l *LocalFuncExecter) Name() string {
	return "local"
}

func (l *LocalFuncExecter) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecter) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未找到任务执行函数: %s", j.Name)
	}
	return fn(ctx, j)
}

type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.Logger
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.Logger) *Scheduler {
	return &Scheduler{
		execs: make(map[string]Executor),
		svc:   svc,
		// 控制任务数量200个
		limiter: semaphore.NewWeighted(200),
		l:       l,
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

// 调度器
func (s *Scheduler) Scheduler(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// main 函数退出调度循环
			return nil
		}
		// 限制并发数，acquire(ctx,1) 的意思是获取一个信号量，如果没有信号量，会阻塞
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 抢占任务
		dbCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 继续下一轮
			s.l.Error("抢占任务失败", logger.Error(err))
		}

		exec, ok := s.execs[j.Executor]
		if !ok {
			s.l.Error("未找到执行器", logger.String("executor", j.Executor))
			continue
		}

		// 接下来就是执行任务，异步执行任务，不阻塞主流程
		go func() {
			// 执行完毕后释放任务
			s.limiter.Release(1)
			defer func() {
				er := j.CancelFunc()
				if er != nil {
					s.l.Error("释放任务失败", logger.Error(er), logger.Int64("job_id", j.Id))
				}
			}()
			// 执行任务，如果失败，记录日志
			er := exec.Exec(ctx, j)
			if er != nil {
				s.l.Error("执行任务失败", logger.Error(er))
			}
			// 考虑下一次调度
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				s.l.Error("设置下次执行时间失败", logger.Error(er), logger.Int64("job_id", j.Id))
			}
		}()
	}
}
