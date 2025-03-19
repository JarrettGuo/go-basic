package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	initPrometheus()
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	// 启动定时任务
	app.cron.Start()

	server := app.server
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})
	// 启动服务器
	server.Run(":8080")

	ctx := app.cron.Stop()
	// 超时强制退出，防止有些任务执行时间过长
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func initViper() {
	// 配置文件名字，不包含文件拓展名
	viper.SetConfigName("dev")
	// 配置用的文件格式
	viper.SetConfigType("yaml")
	// 当前工作目录下的 config 子目录
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
