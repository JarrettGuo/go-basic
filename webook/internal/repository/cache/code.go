package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = fmt.Errorf("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = fmt.Errorf("验证码验证次数过多")
	ErrUnknownForCode         = errors.New("未知错误")
)

//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type RedisCodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
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

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForCode
	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
