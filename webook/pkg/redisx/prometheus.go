package redisx

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	return &PrometheusHook{
		vector: prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"}),
	}
}

// 这个函数是用来实现redis.DialHook接口的
func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// 通过这个函数，我们可以在redis.Dial的时候做一些事情，比如说监控连接耗时
func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		var err error
		defer func() {
			duration := time.Since(start).Milliseconds()
			keyExists := err == redis.Nil
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

// 通过这个函数，我们可以在redis管道执行的时候做一些事情，比如说监控管道执行的耗时
func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
