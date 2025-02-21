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
}
