package startup

import (
	"go-basic/webook/internal/service/oauth2/wechat"
	"go-basic/webook/pkg/logger"
)

func InitWechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "")
}
