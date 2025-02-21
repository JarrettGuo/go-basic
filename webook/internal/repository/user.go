package repository

import (
	"context"
	"database/sql"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/cache"
	"go-basic/webook/internal/repository/dao"
	"time"
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
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
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

	user := r.entityToDomain(ue)
	// 保存到缓存
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 日志记录
		}
	}()
	return user, nil
}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Ctime:    u.Ctime.UnixMilli(),
	}
}
