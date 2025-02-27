package web

import (
	"bytes"
	"errors"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/service"
	svcmocks "go-basic/webook/internal/service/mocks"
	"go-basic/webook/internal/web/jwt"
	jwtmocks "go-basic/webook/internal/web/jwt/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler)

		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)

				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "12312121@qq.com",
					Password: "123456789",
				}).Return(nil)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
					{
						"email": "12312121@qq.com",
						"password": "123456789",
						"confirmPassword": "123456789"
					}`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数错误, bind失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `{
				invalid json format
			}`,
			wantCode: http.StatusBadRequest,
			wantBody: "",
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
					{
						"email": "12312121",
						"password": "123456789",
						"confirmPassword": "123456789"
					}`,
			wantCode: http.StatusOK,
			wantBody: "邮箱格式错误",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "12345",
						"confirmPassword": "12345"
					}`,
			wantCode: http.StatusOK,
			wantBody: "密码格式错误",
		},
		{
			name: "两次密码不一致",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
					{
						"email": "12312121@qq.com",
						"password": "123456789",
						"confirmPassword": "12345678"
					}`,
			wantCode: http.StatusOK,
			wantBody: "两次密码不一致",
		},
		{
			name: "邮箱已存在",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456789",
				}).Return(service.ErrUserDuplicateEmail)
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "123456789",
					"confirmPassword": "123456789"
				}`,
			wantCode: http.StatusOK,
			wantBody: "邮箱已存在",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwthdl := jwtmocks.NewMockHandler(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456789",
				}).Return(errors.New("系统错误"))
				return usersvc, codesvc, jwthdl
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "123456789",
					"confirmPassword": "123456789"
				}`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			usersvc, codesvc, jwthdl := tc.mock(ctrl)
			h := NewUserHandler(usersvc, codesvc, jwthdl)
			h.RegisterRoutes(server)

			// 创建一个请求
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 创建一个响应
			resp := httptest.NewRecorder()
			t.Log(resp)

			// http 请求进入 Gin 的入口，这样调用的时候，GIN就会处理这个请求
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginJWT(t *testing.T) {
	type fields struct {
		svc         service.UserService
		emailExp    *regexp.Regexp
		passwordExp *regexp.Regexp
		nicknameExp *regexp.Regexp
		birthdayExp *regexp.Regexp
		descExp     *regexp.Regexp
		phoneExp    *regexp.Regexp
		codeExp     *regexp.Regexp
		codeSvc     service.CodeService
	}
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserHandler{
				svc:         tt.fields.svc,
				emailExp:    tt.fields.emailExp,
				passwordExp: tt.fields.passwordExp,
				nicknameExp: tt.fields.nicknameExp,
				birthdayExp: tt.fields.birthdayExp,
				descExp:     tt.fields.descExp,
				phoneExp:    tt.fields.phoneExp,
				codeExp:     tt.fields.codeExp,
				codeSvc:     tt.fields.codeSvc,
			}
			u.LoginJWT(tt.args.ctx)
		})
	}
}
