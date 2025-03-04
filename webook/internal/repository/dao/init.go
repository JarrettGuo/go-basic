package dao

import (
	"go-basic/webook/internal/repository/dao/article"

	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &SMSAysncReq{}, &article.Article{}, &article.PublishArticle{})
}
