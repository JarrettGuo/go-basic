package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	// 初始化结构体
	type Comfig struct {
		Addr string `yaml:"addr"`
	}
	var cfg Comfig
	// 读取配置文件
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})
	return redisClient
}
