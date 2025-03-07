//go:build wireinject

package startup

import (
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/repository"
	articleRepository "go-basic/webook/internal/repository/article"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	articleDAO "go-basic/webook/internal/repository/dao/article"
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
	InitMongoDB,
	InitSnowflakeNode,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService,
)

var articlSvcProvider = wire.NewSet(
	// repository.NewCachedArticleRepository,
	// cache.NewArticleRedisCache,
	// dao.NewArticleGORMDAO,
	service.NewArticleService,
	articleDAO.NewReaderDAO,
	articleDAO.NewAuthorDAO,
	articleDAO.NewMongoDBDAO,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,
		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCodeRepository,
		articleRepository.NewArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
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

func InitArticleHandler(dao articleDAO.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		articleDAO.NewReaderDAO,
		articleDAO.NewAuthorDAO,
		articleRepository.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return new(web.ArticleHandler)
}
