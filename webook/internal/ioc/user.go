package ioc

import (
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
	"go-basic/webook/pkg/redisx"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// 如果你不想全局使用 zap.NewDevelopment()，而是想在每个服务中自定义日志，那么你可以在 ioc 包中创建一个 InitLogger 方法，用于初始化 zap.Logger。
func InitUserService(repo repository.UserRepository, l logger.Logger) service.UserService {
	return service.NewUserService(repo, l)
}

// 配合 prometheus 使用
func InitUserCache(client *redis.ClusterClient) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(
		prometheus.SummaryOpts{
			Namespace: "webook",
			Subsystem: "redis",
			Name:      "redis_hook",
			Help:      "统计 redis 命令耗时",
			ConstLabels: map[string]string{
				"biz": "user",
			},
		},
	))
	panic("not implemented")
}
