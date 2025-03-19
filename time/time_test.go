package time

import (
	"log"
	"testing"
	"time"

	cron "github.com/robfig/cron/v3"
	"golang.org/x/net/context"
)

// ticker 等间隔循环触发的计时器
func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Log("timeout or cancel")
			return
		case now := <-ticker.C:
			t.Log(now.Unix())
		}
	}
}

// timer 定时执行的计时器
func TestTimer(t *testing.T) {
	timber := time.NewTimer(time.Second)
	defer timber.Stop()
	for now := range timber.C {
		t.Log(now.Unix())
	}
}

func TestCronExpression(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	expr.AddJob("@every 1s", myJob{})
	expr.AddFunc("@every 3s", func() {
		t.Log("我也运行了")
	})
	// 开始运行
	expr.Start()
	// 模拟运行10秒
	time.Sleep(time.Second * 10)
	// 发出停止信号，expr不会调度新的任务，也不会中断正在运行的任务
	stop := expr.Stop()
	t.Log("停止了")
	// 等待所有任务完成
	<-stop.Done()
	t.Log("停止完成")
}

type myJob struct{}

func (m myJob) Run() {
	log.Println("我运行了")
}
