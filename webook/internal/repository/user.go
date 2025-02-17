package repository

import (
	"context"
	"go-basic/webook/internal/domain"
	"go-basic/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
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
	// 先从 cache 中查找
	// 如果 cache 中没有，再从数据库中查找
	user, err := r.dao.FindById(ctx, Id)
	if err != nil {
		return domain.User{}, err
	}
	// 找到后，将数据写入 cache
	return domain.User{
		Email:    user.Email,
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		Desc:     user.Desc,
		Password: user.Password,
	}, nil
}
