package service

import (
	"context"
	"errors"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("账户或密码错误")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo  repository.UserRepository
	redis *redis.Client
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找到用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG 日志
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateNonZeroFields(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	// 先从缓存获得 user 信息
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先找到用户
	u, err := svc.repo.FindByPhone(ctx, phone)
	// 判断有没有用户
	if err != repository.ErrUserNotFound {
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil {
		return u, err
	}
	// 再次查找，会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}
