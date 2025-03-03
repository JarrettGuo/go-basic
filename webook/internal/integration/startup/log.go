package startup

import "go-basic/webook/pkg/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
