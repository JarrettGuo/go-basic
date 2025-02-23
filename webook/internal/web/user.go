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

var (
	ErrCodeSendTooMany        = service.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = service.ErrCodeVerifyTooManyTimes
)

const biz = "login"

type UserHandler struct {
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

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	const (
		emailRegexPattern = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
		passwordPattern   = `^[a-zA-Z0-9_-]{6,18}$`
		nicknamePattern   = `^.{1,20}$`
		birthdayPattern   = `^\d{4}-\d{2}-\d{2}$`
		descPattern       = `^.{0,200}$`
		phonePattern      = `^1[3-9]\d{9}$`
		codePattern       = `^\d{6}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordPattern, regexp.None)
	nicknameExp := regexp.MustCompile(nicknamePattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayPattern, regexp.None)
	descExp := regexp.MustCompile(descPattern, regexp.None)
	phoneExp := regexp.MustCompile(phonePattern, regexp.None)
	codeExp := regexp.MustCompile(codePattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		nicknameExp: nicknameExp,
		birthdayExp: birthdayExp,
		descExp:     descExp,
		phoneExp:    phoneExp,
		codeExp:     codeExp,
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
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 正则表达式验证手机号
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	// 正则表达式验证验证码
	ok, err = u.codeExp.MatchString(req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	ok, err = u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误",
		})
		return
	}
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if err = u.setJWTToken(user.Id, ctx); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "验证码校验成功",
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 正则表达式验证手机号
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "手机号输入有误",
		})
		return
	}
	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "发送太频繁，请稍后再试",
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
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
	if err = u.setJWTToken(user.Id, ctx); err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
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
	uc := ctx.MustGet("user").(UserClaims)
	err = u.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Desc:     req.Desc,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "修改成功",
	})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email string
	}
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)
	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: Profile{
			Email: user.Email,
		},
	})
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
		Birthday string
		Desc     string
	}
	uc := ctx.MustGet("user").(UserClaims)
	user, err := u.svc.Profile(ctx, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Birthday: user.Birthday,
			Desc:     user.Desc,
		},
	})
}

func (*UserHandler) setJWTToken(uid int64, ctx *gin.Context) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("5131ee22610a224ca4e0869375383995"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明自定义字段
	Uid       int64
	UserAgent string
}
