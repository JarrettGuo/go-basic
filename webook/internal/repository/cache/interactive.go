package cache

import (
	"context"
	_ "embed"
	"fmt"
	"go-basic/webook/internal/domain"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
}

type InteractiveRedisCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

func (c *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt, []string{c.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func (c *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	return c.client.Eval(ctx, luaIncrCnt, []string{c.key(biz, id)}, fieldLikeCnt, 1).Err()
}

func (c *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	return c.client.Eval(ctx, luaIncrCnt, []string{c.key(biz, id)}, fieldLikeCnt, -1).Err()
}

func (c *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error {
	return c.client.Eval(ctx, luaIncrCnt, []string{c.key(biz, id)}, fieldCollectCnt, 1).Err()
}

func (c *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	// HGetAll 当 key 不存在时，返回空 map
	data, err := c.client.HGetAll(ctx, c.key(biz, id)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(data) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}

	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	return domain.Interactive{
		CollectCnt: collectCnt,
		ReadCnt:    readCnt,
		LikeCnt:    likeCnt,
	}, nil
}

func (c *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error {
	key := c.key(biz, bizId)
	err := c.client.HSet(ctx, key, fieldCollectCnt, res.CollectCnt,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt,
	).Err()
	if err != nil {
		return err
	}
	return c.client.Expire(ctx, key, time.Minute*15).Err()
}

func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
