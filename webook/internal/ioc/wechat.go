package ioc

import (
	"go-basic/webook/internal/service/oauth2/wechat"
	"os"
)

func InitOAuth2WechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到 WECHAT_APP_ID")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appKey)
}
