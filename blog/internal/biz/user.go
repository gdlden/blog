package biz

import (
	"context"
	"errors"

	userutil "blog/internal/utils"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

var (
// ErrUserNotFound is user not found.
// ErrUserNotFound = errors.NotFound("USER_NOT_FOUND", "user not found")
)

// User is a User model.
type User struct {
	ID          string
	Username    string
	Password    string
	PhoneNumber string
	Email       string
	Avatar      string
}

// UserRepo is a User repo interface.
type UserRepo interface {
	Create(context.Context, *User) error
	Update(context.Context, *User) error
	Delete(context.Context, string) error
	GetByID(context.Context, string) (*User, error)
	List(context.Context) ([]*User, error)
	GetUser(context.Context, string, string) (*User, error)
	GetUserForRegister(*User) *User // 用于注册时检查用户名是否已存在
	UpdateAvatar(*User) (bool, error)
}

// UserUsecase is a User usecase.
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUsecase new a User usecase.
func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *UserUsecase) Create(ctx context.Context, u *User) error {

	repeatUser := uc.repo.GetUserForRegister(u)
	if repeatUser != nil {
		return errors.New("用户名已存在")
	}
	userPwd := u.Password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(userPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPwd)
	uer := uc.repo.Create(ctx, u)
	return uer
}

func (uc *UserUsecase) Update(ctx context.Context, u *User) error {
	return uc.repo.Update(ctx, u)
}
func (uc *UserUsecase) UpdateAvatar(ctx context.Context, u *User) (bool, error) {
	currentUserId, err2 := userutil.CurrentUserId(ctx)
	if err2 != nil {
		return false, err2
	}
	flag, err := uc.repo.UpdateAvatar(&User{
		ID:     currentUserId,
		Avatar: u.Avatar,
	})
	return flag, err
}

func (uc *UserUsecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *UserUsecase) Get(ctx context.Context, id string) (*User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) List(ctx context.Context) ([]*User, error) {
	return uc.repo.List(ctx)
}
func (uc *UserUsecase) GetUser(ctx context.Context, username string, password string) (*User, error) {
	return uc.repo.GetUser(ctx, username, password)
}
