package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/cache"
	cachemocks "go-basic/webook/internal/repository/cache/mocks"
	"go-basic/webook/internal/repository/dao"
	daomocks "go-basic/webook/internal/repository/dao/mocks"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
)

func TestCacheUserRepository_FindById(t *testing.T) {
	now := time.Now()
	// 将时间戳转换为毫秒
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx      context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，但查询数据库成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{
						Id: 123,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Phone: sql.NullString{
							String: "12345678901",
							Valid:  true,
						},
						Ctime: now.UnixMilli(),
						Utime: now.UnixMilli(),
					}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Phone:    "12345678901",
					Password: "123456",
					Ctime:    now,
				}).Return(nil)
				return d, c
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Phone:    "12345678901",
				Password: "123456",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Phone:    "12345678901",
					Password: "123456",
					Ctime:    now,
				}, nil)
				d := daomocks.NewMockUserDAO(ctrl)
				return d, c
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Phone:    "12345678901",
				Password: "123456",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，DB查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{}, errors.New("mock db error"))
				return d, c
			},
			ctx:      context.Background(),
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("mock db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := NewUserRepository(tc.mock(ctrl))
			user, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
