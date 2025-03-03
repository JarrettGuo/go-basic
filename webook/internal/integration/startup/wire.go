//go:build wireinject

package startup

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

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis,
	InitDB,
	InitLogger,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService,
)

// var articlSvcProvider = wire.NewSet(
// 	repository.NewCachedArticleRepository,
// 	cache.NewArticleRedisCache,
// 	dao.NewArticleGORMDAO,
// 	service.NewArticleService
// )

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		// articlSvcProvider,
		// cache 部分
		cache.NewCodeCache,
		dao.NewArticleDAO,

		// repository 部分
		repository.NewCodeRepository,
		repository.NewArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		service.NewArticleService,
		// InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ioc.InitOAuth2WechatService,
		ioc.NewWechatHandlerConfig,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		// userSvcProvider,
		repository.NewArticleRepository,
		// cache.NewArticleRedisCache,
		dao.NewArticleDAO,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}
