package web

import "go-basic/webook/internal/domain"

// 对应前端的文章数据
type ArticleVO struct {
	Id       int64
	Title    string
	Abstract string
	Content  string
	Author   string
	Status   uint8
	Ctime    string
	Utime    string
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type DetailReq struct {
	Id int64 `json:"id"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
