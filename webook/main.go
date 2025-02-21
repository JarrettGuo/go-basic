package main

import (
	"go-basic/webook/internal/repository"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/service/sms/memory"
	"go-basic/webook/internal/web"
	"go-basic/webook/internal/web/middleware"
	"go-basic/webook/pkg/ginx/middlewares/ratelimit"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/redis/go-redis/v9"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	rdb, server := initWebServer()
	u := initUser(db, rdb)
	u.RegisterRoutes(server)

	server.Run(":8080")
}

func initWebServer() (*redis.Client, *gin.Engine) {
	server := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Authorization", "Content-Type"},
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许携带 cookie
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("5131ee22610a224ca4e0869375383995"), []byte("6131ee22610a224ca4e0869375383995"))
	// if err != nil {
	// 	panic(err)
	// }
	store := memstore.NewStore([]byte("5131ee22610a224ca4e0869375383995"), []byte("6131ee22610a224ca4e0869375383995"))
	server.Use(sessions.Sessions("webook", store))

	// cookie 中间件，登录校验
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		Build())
	return redisClient, server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(rdb, time.Minute*15)
	repo := repository.NewUserRepository(ud, userCache)
	svc := service.NewUserService(repo)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}
