package data

import (
	"blog/internal/biz"
	"blog/internal/model"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) Create(ctx context.Context, u *biz.User) error {
	user := &model.User{
		ID:          u.ID,
		Username:    u.Username,
		Password:    u.Password,
		PhoneNumber: u.PhoneNumber,
		Email:       u.Email,
	}
	return r.data.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) Update(ctx context.Context, u *biz.User) error {
	user := &model.User{
		ID:          u.ID,
		Username:    u.Username,
		Password:    u.Password,
		PhoneNumber: u.PhoneNumber,
		Email:       u.Email,
	}
	return r.data.db.WithContext(ctx).Save(user).Error
}
func (r *userRepo) UpdateAvatar(u *biz.User) (bool, error) {
	var user = model.User{
		ID:     u.ID,
		Avatar: u.Avatar,
	}
	err := r.data.db.Model(&user).Update("avatar", u.Avatar).Error
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*biz.User, error) {
	var user model.User
	if err := r.data.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		ID:          user.ID,
		Username:    user.Username,
		Password:    user.Password,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}

func (r *userRepo) List(ctx context.Context) ([]*biz.User, error) {
	var users []*model.User
	if err := r.data.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}

	result := make([]*biz.User, 0, len(users))
	for _, user := range users {
		result = append(result, &biz.User{
			ID:          user.ID,
			Username:    user.Username,
			Password:    user.Password,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		})
	}
	return result, nil
}
func (r *userRepo) GetUser(ctx context.Context, username string, password string) (*biz.User, error) {
	var user model.User
	if err := r.data.db.WithContext(ctx).Where("username=?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		ID:          user.ID,
		Username:    user.Username,
		Password:    user.Password,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}
func (r *userRepo) GetUserForRegister(u *biz.User) *biz.User {
	var user model.User
	if err := r.data.db.Where("username = ?", u.Username).First(&user).Error; err == nil {
		return &biz.User{
			ID:          user.ID,
			Username:    user.Username,
			Password:    user.Password,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		}
	}

	if u.PhoneNumber != "" {
		var user2 model.User
		if err := r.data.db.Where("phone_number = ?", u.PhoneNumber).First(&user2).Error; err == nil {
			return &biz.User{
				ID:          user2.ID,
				Username:    user2.Username,
				Password:    user2.Password,
				PhoneNumber: user2.PhoneNumber,
				Email:       user2.Email,
			}
		}
	}

	if u.Email != "" {
		var user3 model.User
		if err := r.data.db.Where("email = ?", u.Email).First(&user3).Error; err == nil {
			return &biz.User{
				ID:          user3.ID,
				Username:    user3.Username,
				Password:    user3.Password,
				PhoneNumber: user3.PhoneNumber,
				Email:       user3.Email,
			}
		}
	}

	return nil
}
