package wechat

import (
	"context"
	"os"
	"testing"
)

func Test_service_manual_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到 WECHAT_APP_ID")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到 WECHAT_APP_SECRET")
	}
	svc := NewService(appId, appKey)
	svc.VerifyCode(context.Background(), "code")
}
