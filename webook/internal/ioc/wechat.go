package ioc

import (
	"go-basic/webook/config"
	"go-basic/webook/internal/service/oauth2/wechat"
	"os"

	"github.com/spf13/viper"
)

func InitOAuth2WechatService() wechat.Service {
	type Config struct {
		WECCHAT_APP_ID    string `yaml:"appid"`
		WECHAT_APP_SECRET string `yaml:"secret"`
	}
	var cfg Config
	err := viper.UnmarshalKey("WechatApp", &cfg)
	if err != nil {
		panic(err)
	}
	appId := cfg.WECCHAT_APP_ID
	if envAppID, ok := os.LookupEnv("WECHAT_APP_ID"); ok {
		appId = envAppID
	}

	appKey := cfg.WECHAT_APP_SECRET
	if envAppKey, ok := os.LookupEnv("WECHAT_APP_SECRET"); ok {
		appKey = envAppKey
	}

	return wechat.NewService(appId, appKey)
}

func NewWechatHandlerConfig() config.StateConfig {
	type Config struct {
		Secure   bool   `yaml:"secure"`
		StateKey string `yaml:"stateKey"`
	}
	var cfg Config
	err := viper.UnmarshalKey("OAuth2", &cfg)
	if err != nil {
		panic(err)
	}
	return config.StateConfig{
		Secure:   cfg.Secure,
		StateKey: []byte(cfg.StateKey),
	}
}
