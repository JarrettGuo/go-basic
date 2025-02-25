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

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, Id int64) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	// 保存到数据库
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) FindById(ctx context.Context, Id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, Id)
	// 缓存中有数据
	if err == nil {
		return u, nil
	}

	// 缓存中没有数据
	ue, err := r.dao.FindById(ctx, Id)
	if err != nil {
		return domain.User{}, err
	}

	user := r.entityToDomain(ue)
	// 保存到缓存
	// go func() {
	// 	err = r.cache.Set(ctx, u)
	// 	if err != nil {
	// 		// 日志记录
	// 	}
	// }()
	_ = r.cache.Set(ctx, user)
	return user, nil
}

func (r *CacheUserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	return r.dao.UpdateById(ctx, r.domainToEntity(user))
}

func (r *CacheUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CacheUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenID.String,
			UnionID: u.WechatUnionID.String,
		},
		Ctime: time.UnixMilli(u.Ctime),
	}
}

func (r *CacheUserRepository) domainToEntity(u domain.User) dao.User {
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
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}
