package ratelimit

import "context"

type Limiter interface {
	// Limited 有没有触发限流。 key 就是限流对象，bool 表示是否触发限流，error 表示错误信息
	Limit(ctx context.Context, key string) (bool, error)
}
