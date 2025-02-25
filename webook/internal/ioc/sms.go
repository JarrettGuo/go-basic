package ioc

import (
	"go-basic/webook/internal/service/ratelimit"
	"go-basic/webook/internal/service/sms"
	"go-basic/webook/internal/service/sms/memory"
	limiter "go-basic/webook/pkg/ratelimit"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 换内存还是换短信服务，只需要修改这里
	svc := memory.NewService()
	return ratelimit.NewRatelimitSMSService(svc, limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
}
