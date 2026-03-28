package biz

import (
	"blog/internal/utils"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type DebtDetail struct {
	Id          uint
	DebtId      uint
	PostingDate string
	Principal   decimal.Decimal
	Interest    decimal.Decimal
	Period      uint
}

type DebtDetailRepo interface {
	SaveDb(context.Context, *DebtDetail) (string, error)
	GetByUserIdAndID(context.Context, string, uint) (*DebtDetail, error)
	ListByUserId(context.Context, string) ([]*DebtDetail, error)

	EditDb(context.Context, *DebtDetail)

	DeleteDb(context.Context, uint)
}

type DebtDetailUsecase struct {
	ddRepo DebtDetailRepo
	log    *log.Helper
}

func (u *DebtDetailUsecase) Save(ctx context.Context, b *DebtDetail) (string, error) {
	id, err := u.ddRepo.SaveDb(ctx, b)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (u *DebtDetailUsecase) Get(ctx context.Context, id uint) (*DebtDetail, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return u.ddRepo.GetByUserIdAndID(ctx, userId, id)
}

func (u *DebtDetailUsecase) List(ctx context.Context) ([]*DebtDetail, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return u.ddRepo.ListByUserId(ctx, userId)
}

func NewDeptUseCase(repo DebtDetailRepo, logger log.Logger) *DebtDetailUsecase {
	return &DebtDetailUsecase{
		ddRepo: repo,
		log:    log.NewHelper(logger),
	}
}
