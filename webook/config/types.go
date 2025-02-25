package config

type config struct {
	DB    DBConfig
	Redis RedisConfig
	State StateConfig
	OAuth OAuth2Config
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}

type StateConfig struct {
	Secure   bool
	StateKey []byte
}

type OAuth2Config struct {
	WechatAppID     string
	WechatAppSecret string
}
