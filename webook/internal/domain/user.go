package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Phone    string
	Password string
	Ctime    time.Time
	Nickname string
	Birthday string
	Desc     string
	// 不要组合，万一将来有同名字段，会有问题
	WechatInfo WechatInfo
}
