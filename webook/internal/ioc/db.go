package ioc

import (
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/pkg/logger"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	// 初始化结构体
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	// 读取配置文件
	err := viper.UnmarshalKey("db.mysql", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 慢 SQL 阈值，超过该阈值的 SQL 将被记录
			SlowThreshold: time.Millisecond * 10,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

// gormLoggerFunc 转换为 logger.Logger，以适配 gorm 的日志接口，单接口的方法可以实现，多借口不适用
type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{
		Key:   "args",
		Value: args,
	})
}
