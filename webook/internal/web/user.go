package web

import (
	"fmt"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/service"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	nicknameExp *regexp.Regexp
	birthdayExp *regexp.Regexp
	descExp     *regexp.Regexp
	codeSvc     *service.CodeService
}

func NewUserHandler(svc *service.UserService, codeSvc *service.CodeService) *UserHandler {
	const (
		emailRegexPattern = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
		passwordPattern   = `^[a-zA-Z0-9_-]{6,18}$`
		nicknamePattern   = `^.{1,20}$`
		birthdayPattern   = `^\d{4}-\d{2}-\d{2}$`
		descPattern       = `^.{0,200}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordPattern, regexp.None)
	nicknameExp := regexp.MustCompile(nicknamePattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayPattern, regexp.None)
	descExp := regexp.MustCompile(descPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		nicknameExp: nicknameExp,
		birthdayExp: birthdayExp,
		descExp:     descExp,
		codeSvc:     codeSvc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/login", u.LoginJWT)
	ug.POST("/signup", u.SignUp)
	ug.PUT("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	const biz = "login"
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "发送成功",
	})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 调用 service 层的登录方法
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "账户或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 设置 session
	sess := sessions.Default(ctx)
	// 设置 session 的值
	sess.Set("userId", user.Id)
	// 设置 session 的自定义配置
	sess.Options(sessions.Options{
		// 设置 session 过期时间
		MaxAge: 3600,
	})
	// 保存 session
	sess.Save()

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 调用 service 层的登录方法
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "账户或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 用 JWT 设置登录态
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("5131ee22610a224ca4e0869375383995"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	ctx.String(http.StatusOK, "登录成功")
	return
}

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
	// 调用 service 层的注册方法
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱已存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Desc     string `json:"desc"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 正则表达式验证昵称
	ok, err := u.nicknameExp.MatchString(req.Nickname)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "昵称格式错误")
		return
	}
	// 正则表达式验证生日
	ok, err = u.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "生日格式错误")
		return
	}
	// 正则表达式验证描述
	ok, err = u.descExp.MatchString(req.Desc)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "描述格式错误")
		return
	}
	// 获取 session 中的 userId
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	if userId == nil {
		ctx.String(http.StatusUnauthorized, "请先登录")
		return
	}
	// 调用 service 层的编辑方法
	err = u.svc.Edit(ctx, domain.User{
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Desc:     req.Desc,
		Id:       userId.(int64),
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "编辑成功")
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	// 获取 session 中的 userId
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	if userId == nil {
		ctx.String(http.StatusUnauthorized, "请先登录")
		return
	}
	// 调用 service 层的获取用户信息方法
	user, err := u.svc.Profile(ctx, userId.(int64))
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "用户信息：%+v", user)
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 调用 service 层的获取用户信息方法
	user, err := u.svc.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "用户信息：%+v", user)
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明自定义字段
	Uid       int64
	UserAgent string
}
