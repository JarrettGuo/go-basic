package ratelimit

import (
	"fmt"
	"go-basic/webook/pkg/ratelimit"
	"log"
	"net/http"

	_ "embed"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	prefix   string
	limliter ratelimit.Limiter
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		limliter: limiter,
		prefix:   "ip-limiter",
	}
}

func (b *Builder) Profix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println("limit error:", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limliter.Limit(ctx, key)
}
