package main

import (
	"go-basic/webook/internal/web"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
