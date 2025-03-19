package job

import (
	"go-basic/webook/pkg/logger"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

type CronJobBuilder struct {
	l logger.Logger
	p *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.Logger) *CronJobBuilder {
	p := prometheus.NewSummaryVec(prometheus.SummaryOpts(prometheus.SummaryOpts{
		Namespace: "webook",
		Subsystem: "job",
		Name:      "cron_job",
		Help:      "统计定时任务的执行时间",
	}), []string{"name", "success"})
	prometheus.MustRegister(p)
	return &CronJobBuilder{
		l: l,
		p: p,
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		start := time.Now()
		b.l.Info("开始运行任务", logger.String("job", name))
		var success bool
		defer func() {
			b.l.Debug("任务运行结束", logger.String("job", name))
			duration := time.Since(start).Milliseconds()
			b.p.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run()
		success = err == nil
		if err != nil {
			b.l.Error("运行任务失败", logger.Error(err), logger.String("job", name))
		}
		return nil
	})
}

type cronJobFuncAdapter func() error

func (c cronJobFuncAdapter) Run() {
	c()
}
