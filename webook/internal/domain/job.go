package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Id         int64
	Name       string
	Cron       string
	Executor   string
	Cfg        string
	Version    int
	CancelFunc func() error
}

var parse = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	// 根据cron表达式计算下一次执行时间
	s, _ := parse.Parse(j.Cron)
	return s.Next(time.Now())
}
