package ioc

import (
	"go-basic/webook/internal/web"
	"go-basic/webook/internal/web/middleware"
	"go-basic/webook/pkg/ginx/middlewares/ratelimit"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
