package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱已存在")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	UpdateUserProfile(ctx context.Context, u User) error
}
type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateUserProfile(ctx context.Context, u User) error {
	u.Utime = time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Model(&u).Where("email = ?", u.Email).Updates(&u).Error
	return err
}

// 直接对应数据库表结构
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 唯一索引，可以允许有多个空值，但是不能允许有多个""
	Email    sql.NullString `gorm:"unique"`
	Password string
	Phone    sql.NullString `gorm:"unique"`
	Nickname string
	Birthday string
	Desc     string
	// 时间，存毫秒
	Ctime int64
	Utime int64
}
