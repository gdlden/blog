package data

import (
	"context"
	"time"

	"blog/internal/biz"
	"blog/internal/model"

	"github.com/go-kratos/kratos/v2/log"
)

type verificationRepo struct {
	data *Data
	log  *log.Helper
}

func NewVerificationRepo(data *Data, logger log.Logger) biz.VerificationRepo {
	return &verificationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *verificationRepo) Save(ctx context.Context, vc *biz.VerificationCode) error {
	m := &model.VerificationCode{
		Email:     vc.Email,
		Code:      vc.Code,
		Type:      vc.Type,
		ExpiresAt: vc.ExpiresAt,
	}
	return r.data.db.WithContext(ctx).Create(m).Error
}

func (r *verificationRepo) FindLatestByEmail(ctx context.Context, email string, codeType string) (*biz.VerificationCode, error) {
	var m model.VerificationCode
	if err := r.data.db.WithContext(ctx).
		Where("email = ? AND type = ?", email, codeType).
		Order("created_at desc").
		First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.VerificationCode{
		ID:        m.ID,
		Email:     m.Email,
		Code:      m.Code,
		Type:      m.Type,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *verificationRepo) MarkUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).
		Model(&model.VerificationCode{}).
		Where("id = ?", id).
		Update("used_at", &now).Error
}
