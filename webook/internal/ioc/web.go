package ioc

import (
	"context"
	"go-basic/webook/internal/web"
	ijwt "go-basic/webook/internal/web/jwt"
	"go-basic/webook/internal/web/middleware"
	"go-basic/webook/pkg/ginx/middlewares/logger"
	loggerx "go-basic/webook/pkg/logger"

	"go-basic/webook/pkg/ginx/middlewares/ratelimit"
	ratelimitx "go-basic/webook/pkg/ratelimit"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l loggerx.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
			l.Debug("HTTP Request", loggerx.Field{Key: "al", Value: al})
		}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			Build(),
		ratelimit.NewBuilder(ratelimitx.NewRedisSlidingWindowLimiter(redisClient, time.Second, 100)).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"x-jwt-token, x-refresh-token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
