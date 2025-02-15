package web

import (
	"fmt"
	"net/http"

	regexp "github.com/dlclark/regexp2"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/login", u.Login)
	ug.POST("/signup", u.SignUp)
	ug.PUT("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
}

func NewUserHandler() *UserHandler {
	const (
		emailRegexPattern = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
		passwordPattern   = `^[a-zA-Z0-9_-]{6,18}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordPattern, regexp.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) Login(ctx *gin.Context) {}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	// Bind 方法会根据 content-type 来解析道数据 req 中，如果解析失败会返回 400 错误
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 正则表达式验证邮箱
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}
	// 正则表达式验证密码
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式错误")
		return
	}
	ctx.String(http.StatusOK, "success")
	fmt.Println(req)
	// 数据库操作
}

func (u *UserHandler) Edit(ctx *gin.Context) {}

func (u *UserHandler) Profile(ctx *gin.Context) {}
