package ginx

import (
	"go-basic/webook/internal/web/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T, uc jwt.UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		res, err := fn(ctx, req, ctx.MustGet("user").(jwt.UserClaims))
		if err != nil {

		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
