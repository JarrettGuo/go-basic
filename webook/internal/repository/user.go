package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	// 保存到数据库
	return r.dao.Insert(ctx, dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	})
	// 保存到缓存
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, u domain.User) error {
	// 更新用户信息
	return r.dao.UpdateUserProfile(ctx, dao.User{
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Desc:     u.Desc,
	})
}

func (r *UserRepository) FindById(ctx context.Context, Id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, Id)
	// 缓存中有数据
	if err == nil {
		return u, nil
	}
	// 缓存中没有数据
	if err == cache.ErrKeyNotExist {
		// 去数据库中查找
	}
	ue, err := r.dao.FindById(ctx, Id)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		Password: ue.Password,
	}
	// 保存到缓存
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 日志记录
		}
	}()
	return user, nil
}
