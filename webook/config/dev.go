//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
	State: StateConfig{
		Secure:   false,
		StateKey: []byte("5131ee22610a224ca4e0869375383912"),
	},
	OAuth: OAuth2Config{
		WechatAppID:     "your_dev_app_id",
		WechatAppSecret: "your_dev_app_secret",
	},
}
