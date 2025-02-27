package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type SMSDAO interface {
	Insert(ctx context.Context, req SMSAysncReq) error
}

type SMSAysncReqDAO struct {
	db *gorm.DB
}

func NewSMSAysncReqDAO(db *gorm.DB) SMSDAO {
	return &SMSAysncReqDAO{
		db: db,
	}
}

func (dao *SMSAysncReqDAO) Insert(ctx context.Context, req SMSAysncReq) error {
	now := time.Now().UnixMilli()
	req.Ctime = now
	req.Utime = now
	err := dao.db.WithContext(ctx).Create(&req).Error
	return err
}

func (dao *SMSAysncReqDAO) FindByStatus(ctx context.Context, status int8) ([]SMSAysncReq, error) {
	var resp []SMSAysncReq
	err := dao.db.WithContext(ctx).Where("status = ?", status).Find(&resp).Error
	return resp, err
}

type SMSAysncReq struct {
	Id       int64 `gorm:"primaryKey;autoIncrement"`
	Biz      string
	Args     string
	Numbers  string
	Status   int8 `gorm:"type:tinyint;not null;default:0;comment:状态:0-等待发送,1-发送中,2-发送成功,3-发送失败"`
	RetryCnt int32
	Ctime    int64
	Utime    int64
}
