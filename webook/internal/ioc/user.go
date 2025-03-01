package ioc

import (
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/service"
	"go-basic/webook/pkg/logger"
)

// 如果你不想全局使用 zap.NewDevelopment()，而是想在每个服务中自定义日志，那么你可以在 ioc 包中创建一个 InitLogger 方法，用于初始化 zap.Logger。
func InitUserService(repo repository.UserRepository, l logger.Logger) service.UserService {
	return service.NewUserService(repo, l)
}
