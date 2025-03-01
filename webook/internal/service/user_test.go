package service

import (
	"context"
	"errors"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository"
	repomocks "go-basic/webook/internal/repository/mocks"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"

	"golang.org/x/crypto/bcrypt"
)

func Test_userService_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		ctx      context.Context
		email    string
		password string

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Phone:    "12345678901",
						Email:    "123@qq.com",
						Password: "$2a$10$znmkpsZFBIeHrsDhCeWoLOWOHyP5lHBA.XL1fU8gypMO28JcMb20O",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "123456",
			wantUser: domain.User{
				Phone:    "12345678901",
				Email:    "123@qq.com",
				Password: "$2a$10$znmkpsZFBIeHrsDhCeWoLOWOHyP5lHBA.XL1fU8gypMO28JcMb20O",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "123456",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},
			email:    "123@qq.com",
			password: "123456",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Phone:    "12345678901",
						Email:    "123@qq.com",
						Password: "$2a$10$znmkpsZFBIeHrsDhCeWoLOWOHyP5lHBA.XL1fU8gypMO28JcMb20O",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "12345677",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl), nil)
			u, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}
