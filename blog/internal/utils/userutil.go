package utils

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func CurrentUserId(ctx context.Context) (string, error) {
	token, ok := jwt.FromContext(ctx)
	if !ok {
		return "", errors.New("未登录")
	}
	userInfoMap := token.(jwtv5.MapClaims)
	userId := userInfoMap["userId"]
	return userId.(string), nil
}
