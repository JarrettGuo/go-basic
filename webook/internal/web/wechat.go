package web

import (
	"errors"
	"fmt"
	"go-basic/webook/config"
	"go-basic/webook/internal/service"
	"go-basic/webook/internal/service/oauth2/wechat"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"

	"github.com/gin-gonic/gin"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
	cfg config.StateConfig
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, cfg config.StateConfig) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
		cfg:     cfg,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "获取授权地址失败",
		})
		return
	}
	err = h.setStateCookie(state, ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) setStateCookie(state string, ctx *gin.Context) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString(h.cfg.StateKey)
	if err != nil {
		return fmt.Errorf("生成 token 失败, %w", err)
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "oauth2/wechat/callback", "", h.cfg.Secure, true)
	return nil
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "登录失败",
		})
		return
	}
	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 验证通过，设置JWT Token
	user, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.setJWTToken(user.Id, ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("拿不到 state 的 cookie, %w", err)
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.cfg.StateKey, nil
	})
	if err != nil {
		return fmt.Errorf("token 已经过期了, %w", err)
	}
	if sc.State != state || !token.Valid {
		return errors.New("state 不匹配")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
