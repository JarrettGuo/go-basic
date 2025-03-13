// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	article3 "go-basic/webook/events/article"
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/repository"
	article2 "go-basic/webook/internal/repository/article"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/internal/repository/dao/article"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/web"
	"go-basic/webook/internal/web/jwt"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, handler, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, logger)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler, logger)
	wechatService := ioc.InitOAuth2WechatService()
	stateConfig := ioc.NewWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler, stateConfig)
	articleDAO := article.NewGORMArticleDAO(db)
	readerDAO := article.NewReaderDAO(db)
	authorDAO := article.NewAuthorDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDAO, readerDAO, authorDAO, articleCache, logger)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, logger, producer)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, logger, interactiveService)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventBatchConsumer := article3.NewInteractiveReadEventBatchConsumer(client, interactiveRepository, logger)
	v2 := ioc.InitConsumers(interactiveReadEventBatchConsumer)
	app := &App{
		server:    engine,
		consumers: v2,
	}
	return app
}
