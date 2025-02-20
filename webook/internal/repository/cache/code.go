package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany = fmt.Errorf("send code too frequently")
)

//go:embed lua/set_code.lua
var luaSetCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		// success
		return nil
	case -1:
		// send so frequently
		return ErrCodeSendTooMany
	default:
		// system error
		return errors.New("system error")
	}
}

func (c *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
