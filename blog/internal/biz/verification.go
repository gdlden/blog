package biz

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type VerificationCode struct {
	ID        uint
	Email     string
	Code      string
	Type      string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type VerificationRepo interface {
	Save(ctx context.Context, vc *VerificationCode) error
	FindLatestByEmail(ctx context.Context, email string, codeType string) (*VerificationCode, error)
	MarkUsed(ctx context.Context, id uint) error
}

type VerificationUsecase struct {
	repo  VerificationRepo
	email *EmailSender
	log   *log.Helper
}

func NewVerificationUsecase(repo VerificationRepo, email *EmailSender, logger log.Logger) *VerificationUsecase {
	return &VerificationUsecase{
		repo:  repo,
		email: email,
		log:   log.NewHelper(logger),
	}
}

func (uc *VerificationUsecase) SendCode(ctx context.Context, email string) error {
	// 检查是否在 60 秒内重复发送
	existing, err := uc.repo.FindLatestByEmail(ctx, email, "register")
	if err == nil && existing != nil && time.Since(existing.CreatedAt) < 60*time.Second {
		return errors.New("请 60 秒后再试")
	}

	code := generateCode()
	now := time.Now()
	vc := &VerificationCode{
		Email:     email,
		Code:      code,
		Type:      "register",
		ExpiresAt: now.Add(10 * time.Minute),
	}

	if err := uc.repo.Save(ctx, vc); err != nil {
		return errors.New("保存验证码失败")
	}

	if err := uc.email.SendVerificationCode(email, code); err != nil {
		return err
	}

	return nil
}

func (uc *VerificationUsecase) VerifyCode(ctx context.Context, email string, code string) error {
	vc, err := uc.repo.FindLatestByEmail(ctx, email, "register")
	if err != nil {
		return errors.New("验证码不存在或已过期")
	}
	if vc == nil {
		return errors.New("验证码不存在或已过期")
	}
	if vc.UsedAt != nil {
		return errors.New("验证码已被使用")
	}
	if time.Now().After(vc.ExpiresAt) {
		return errors.New("验证码已过期")
	}
	if vc.Code != code {
		return errors.New("验证码错误")
	}

	if err := uc.repo.MarkUsed(ctx, vc.ID); err != nil {
		return errors.New("标记验证码失败")
	}

	return nil
}
