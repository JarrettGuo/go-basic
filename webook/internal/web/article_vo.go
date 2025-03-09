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
	// 准确的计数
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	// 个人有没有点赞和收藏
	Liked     bool
	Collected bool
	Ctime     string
	Utime     string
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

// 点赞和取消点赞一个请求
type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

// cid 是收藏夹的ID
type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
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
