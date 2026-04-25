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
	ListByUserIdAndDebtId(context.Context, string, uint) ([]*DebtDetail, error)
	EditDb(context.Context, *DebtDetail) error
	DeleteDb(context.Context, string, uint) error
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

func (u *DebtDetailUsecase) ListByDebtId(ctx context.Context, debtId uint) ([]*DebtDetail, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return u.ddRepo.ListByUserIdAndDebtId(ctx, userId, debtId)
}

func (u *DebtDetailUsecase) Edit(ctx context.Context, b *DebtDetail) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	_, err = u.ddRepo.GetByUserIdAndID(ctx, userId, b.Id)
	if err != nil {
		return err
	}
	return u.ddRepo.EditDb(ctx, b)
}

func (u *DebtDetailUsecase) Delete(ctx context.Context, id uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	return u.ddRepo.DeleteDb(ctx, userId, id)
}

func NewDeptUseCase(repo DebtDetailRepo, logger log.Logger) *DebtDetailUsecase {
	return &DebtDetailUsecase{
		ddRepo: repo,
		log:    log.NewHelper(logger),
	}
}
