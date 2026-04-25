package service

import (
	pb "blog/api/user/v1"
	"blog/internal/biz"
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	pb.UnimplementedUserServer
	uc  *biz.UserUsecase
	log *log.Helper
}

func NewUserService(uc *biz.UserUsecase, logger log.Logger) *UserService {
	return &UserService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	user := &biz.User{
		ID:          uuid.New().String(),
		Username:    req.Username,
		Password:    req.Password, // 实际应用中需要加密
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	}

	if err := s.uc.Create(ctx, user); err != nil {
		return nil, err
	}

	return &pb.CreateUserReply{
		UserId:      user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	user := &biz.User{
		ID:          req.UserId,
		Username:    req.Username,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	}

	if err := s.uc.Update(ctx, user); err != nil {
		return nil, err
	}

	return &pb.UpdateUserReply{
		UserId:      user.ID,
		Username:    user.Username,
		Password:    user.Password,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}
func (s *UserService) UpdateMyAvatar(ctx context.Context, req *pb.UpdateAvatarRequest) (*pb.UpdateAvatarReply, error) {
	flag, _ := s.uc.UpdateAvatar(ctx, &biz.User{
		Avatar: req.Avatar,
	})
	return &pb.UpdateAvatarReply{
		Flag: flag,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	err := s.uc.Delete(ctx, req.UserId)
	if err != nil {
		return &pb.DeleteUserReply{Flag: false}, err
	}
	return &pb.DeleteUserReply{Flag: true}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	user, err := s.uc.Get(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserReply{
		UserId:      user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	users, err := s.uc.List(ctx)
	if err != nil {
		return nil, err
	}

	reply := &pb.ListUserReply{
		UserList: make([]*pb.UserCommom, 0, len(users)),
	}

	for _, user := range users {
		reply.UserList = append(reply.UserList, &pb.UserCommom{
			UserId:      user.ID,
			Username:    user.Username,
			Password:    user.Password,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		})
	}

	return reply, nil
}
func (s *UserService) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.LoginReply, error) {
	user, error := s.uc.GetUser(ctx, req.Username, req.Password)
	if error != nil {
		return nil, error
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"userId":   user.ID,
	})
	if token == nil {
		return nil, biz.ErrUserNotFound
	}

	// Use a secret key for signing (replace with your own secure key)
	secretKey := []byte("dfsdsjikldsfkdfjdkls")
	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		Code:  "200",
		Msg:   "Login successful",
		Token: tokenStr,
		User: &pb.UserCommom{
			UserId:      user.ID,
			Username:    user.Username,
			Password:    "",
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		},
	}, nil
}
