package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("post", func(c *gin.Context) {
		c.String(http.StatusOK, "post")
	})

	// 参数路由，:name 为参数
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})

	// 通配符路由，*name 为通配符
	r.GET("/html/*.file", func(c *gin.Context) {
		file := c.Param(".file")
		c.String(http.StatusOK, "This is %s", file)
	})

	// 查询参数
	r.GET("/order", func(c *gin.Context) {
		id := c.Query("id")
		c.String(http.StatusOK, "The id is %s", id)
	})

	r.Run()
}
