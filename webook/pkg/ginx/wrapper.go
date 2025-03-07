package ginx

import (
	"go-basic/webook/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var L logger.Logger

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c C
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok = val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 接下来的逻辑
		res, err := fn(ctx, c)
		if err != nil {
			L.Error("处理业务逻辑出错",
				// 请求的具体路径
				logger.String("path", ctx.Request.URL.Path),
				// 业务逻辑的具体路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyAndToken[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		var c C
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok = val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 接下来的逻辑
		res, err := fn(ctx, req, c)
		if err != nil {
			L.Error("处理业务逻辑出错",
				// 请求的具体路径
				logger.String("path", ctx.Request.URL.Path),
				// 业务逻辑的具体路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		// 接下来的逻辑
		res, err := fn(ctx, req)
		if err != nil {
			L.Error("处理业务逻辑出错",
				// 请求的具体路径
				logger.String("path", ctx.Request.URL.Path),
				// 业务逻辑的具体路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
