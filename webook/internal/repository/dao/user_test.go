package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/assert/v2"
	"github.com/go-sql-driver/mysql"

	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 只需要INSERT到users表就可以，不需要检查具体的数据
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			ctx: context.Background(),
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 只需要INSERT到users表就可以，不需要检查具体的数据
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysql.MySQLError{
					Number: 1062,
				})
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrUserDuplicateEmail,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 只需要INSERT到users表就可以，不需要检查具体的数据
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(errors.New("mock db error"))
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: errors.New("mock db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: tc.mock(t),
				// 跳过初始化版本
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 禁用自动ping
				DisableAutomaticPing: true,
				// 禁用默认事务
				SkipDefaultTransaction: true,
			})
			d := NewUserDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
