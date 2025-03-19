//go:build wireinject

package main

import (
	artEvt "go-basic/webook/events/article"
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/repository"
	artRepo "go-basic/webook/internal/repository/article"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	articleDAO "go-basic/webook/internal/repository/dao/article"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/web"
	ijwt "go-basic/webook/internal/web/jwt"

	"github.com/google/wire"
)

var rankingServiceSet = wire.NewSet(
	repository.NewRankingRepository,
	cache.NewRankingRedisCache,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		ioc.NewWechatHandlerConfig,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitConsumers,

		rankingServiceSet,
		ioc.InitJob,
		ioc.InitRankingJob,

		// consumer
		artEvt.NewKafkaProducer,
		artEvt.NewInteractiveReadEventBatchConsumer,

		dao.NewUserDAO,
		dao.NewGORMInteractiveDAO,
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,
		cache.NewInteractiveRedisCache,
		articleDAO.NewGORMArticleDAO,
		articleDAO.NewReaderDAO,
		articleDAO.NewAuthorDAO,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedInteractiveRepository,
		artRepo.NewArticleRepository,

		ioc.InitOAuth2WechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,
		ioc.InitSMSService,
		ijwt.NewRedisJWTHandler,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitWebServer,
		ioc.InitMiddlewares,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
