package main

import (
	"go-basic/webook/internal/web"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		// 是否允许携带 cookie
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
