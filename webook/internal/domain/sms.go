package domain

import "time"

type SMS struct {
	Id       int64
	Biz      string
	Args     []string
	Numbers  []string
	Status   int8
	RetryCnt int32
	Ctime    time.Time
}
