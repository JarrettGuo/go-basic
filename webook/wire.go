//go:build wireinject

package main

import (
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/web"
	ijwt "go-basic/webook/internal/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		ioc.NewWechatHandlerConfig,
		ioc.InitLogger,

		dao.NewUserDAO,
		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		ioc.InitOAuth2WechatService,
		service.NewUserService,
		service.NewCodeService,
		ioc.InitSMSService,
		ijwt.NewRedisJWTHandler,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
