// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package integration

import (
	"github.com/gin-gonic/gin"
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/web"
	"go-basic/webook/internal/web/jwt"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitMiddlewares(cmdable, handler)
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitOAuth2WechatService()
	stateConfig := ioc.NewWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, stateConfig, handler)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler)
	return engine
}
