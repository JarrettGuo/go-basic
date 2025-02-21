package ioc

import (
	"go-basic/webook/internal/service/sms"
	"go-basic/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 换内存还是换短信服务，只需要修改这里
	return memory.NewService()
}
