package ioc

import (
	"go-basic/webook/config"
	"go-basic/webook/internal/service/oauth2/wechat"
	"os"
)

func InitOAuth2WechatService() wechat.Service {
	appId := config.Config.OAuth.WechatAppID
	if envAppID, ok := os.LookupEnv("WECHAT_APP_ID"); ok {
		appId = envAppID
	}

	appKey := config.Config.OAuth.WechatAppSecret
	if envAppKey, ok := os.LookupEnv("WECHAT_APP_SECRET"); ok {
		appKey = envAppKey
	}

	return wechat.NewService(appId, appKey)
}

func NewWechatHandlerConfig() config.StateConfig {
	return config.Config.State
}
