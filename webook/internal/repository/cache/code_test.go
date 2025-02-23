package cache

import (
	"context"
	_ "embed"
	"errors"
	"go-basic/webook/internal/repository/cache/redismocks"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "验证码存储成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{"phone_code:login:12345678901"},
					[]string{"123456"},
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678901",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("mock redis error"))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{"phone_code:login:12345678901"},
					[]string{"123456"},
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678901",
			code:    "123456",
			wantErr: errors.New("mock redis error"),
		},
		{
			name: "发送验证码太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{"phone_code:login:12345678901"},
					[]string{"123456"},
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678901",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(999))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{"phone_code:login:12345678901"},
					[]string{"123456"},
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678901",
			code:    "123456",
			wantErr: errors.New("system error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
