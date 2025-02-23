package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"go-basic/webook/internal/ioc"
	"go-basic/webook/internal/web"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string
		// 准备数据
		before func(t *testing.T)
		// 验证数据
		after    func(t *testing.T)
		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送验证码成功",
			// redis不需要准备数据
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13812345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			reqBody: `{
				"phone": "13812345678"
			}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "发送太频繁",
			// redis不需要准备数据
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err := rdb.Set(ctx, "phone_code:login:13812345678", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13812345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "13812345678"
			}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg:  "发送太频繁，请稍后再试",
				Code: 3,
			},
		},
		{
			name: "系统错误",
			// redis不需要准备数据
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err := rdb.Set(ctx, "phone_code:login:13812345678", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13812345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "13812345678"
			}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg:  "系统错误",
				Code: 5,
			},
		},
		{
			name: "手机号错误",
			// redis不需要准备数据
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			reqBody: `{
				"phone": "123456"
			}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg:  "手机号输入有误",
				Code: 4,
			},
		},
		{
			name: "数据格式错误",
			// redis不需要准备数据
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			reqBody: `{
				"phone":,
			}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			// 创建一个请求
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 创建一个响应
			resp := httptest.NewRecorder()
			t.Log(resp)

			// http 请求进入 Gin 的入口，这样调用的时候，GIN就会处理这个请求
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)

			if resp.Code != http.StatusOK {
				return
			}
			// 解析响应
			var webRes web.Result
			// 将响应的body解析为json
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tc.after(t)
		})
	}
}
