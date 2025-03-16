package ioc

import (
	"go-basic/webook/internal/repository/dao"
	"go-basic/webook/pkg/logger"
	"time"

	promsdk "github.com/prometheus/client_golang/prometheus"

	"gorm.io/plugin/prometheus"

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
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running", "Threads_connected"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 监控查询耗时
	pcb := newCallbacks()
	pcb.registerAll(db)

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

// gormLoggerFunc 转换为 logger.Logger，以适配 gorm 的日志接口，单接口的方法可以实现，多接口不适用
type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{
		Key:   "args",
		Value: args,
	})
}

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func newCallbacks() *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: "webook",
		Subsystem: "gorm",
		Name:      "query_duration",
		Help:      "统计 gorm 查询耗时",
		Objectives: map[float64]float64{
			0.5:   0.05,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	},
		// 监控了 table 和 type 两个标签，分别代表了表名和操作类型
		[]string{"type", "table"},
	)
	pcb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return pcb
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues("create", table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}

func (pcb *Callbacks) registerAll(db *gorm.DB) {
	// 作用于 insert 语句，在 insert 之前执行
	err := db.Callback().Create().Before("*").Register("prometheus_create_before", pcb.before())
	if err != nil {
		panic(err)
	}
	// 作用于 insert 语句，在 insert 之后执行
	err = db.Callback().Create().After("*").Register("prometheus_create_after", pcb.after("create"))
	if err != nil {
		panic(err)
	}
	// 作用于 update 语句，在 update 之前执行
	err = db.Callback().Update().Before("*").Register("prometheus_update_before", pcb.before())
	if err != nil {
		panic(err)
	}
	// 作用于 update 语句，在 update 之后执行
	err = db.Callback().Update().After("*").Register("prometheus_update_after", pcb.after("update"))
	if err != nil {
		panic(err)
	}
	// 作用于 delete 语句，在 delete 之前执行
	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", pcb.before())
	if err != nil {
		panic(err)
	}
	// 作用于 delete 语句，在 delete 之后执行
	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", pcb.after("delete"))
	if err != nil {
		panic(err)
	}
	// 作用于 row 语句，在 row 之前执行，row 语句是查询单条记录的语句
	err = db.Callback().Row().Before("*").Register("prometheus_row_before", pcb.before())
	if err != nil {
		panic(err)
	}
	// 作用于 row 语句，在 row 之后执行
	err = db.Callback().Row().After("*").Register("prometheus_row_after", pcb.after("row"))
	if err != nil {
		panic(err)
	}
	// 作用于 raw 语句，在 raw 之前执行，raw 语句是直接执行 SQL 语句的方法
	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before", pcb.before())
	if err != nil {
		panic(err)
	}
	// 作用于 raw 语句，在 raw 之后执行
	err = db.Callback().Raw().After("*").Register("prometheus_raw_after", pcb.after("raw"))
	if err != nil {
		panic(err)
	}
}
