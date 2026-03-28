package biz

import (
	"blog/internal/utils"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type Debt struct {
	Id          int64
	Name        string
	BankName    string
	BankAccount string
	ApplyTime   string
	EndTime     string
	Amount      decimal.Decimal
	Tenor       decimal.Decimal
	Status      int
	Remark      string
	Apr         decimal.Decimal
	Fee         decimal.Decimal
	UserId      string
}

type DebtListQuery struct {
	Page     int64
	PageSize int64
	Name     string
	BankName string
	Status   *int
}

type DebtRepo interface {
	Save(context.Context, *Debt) (uint, error)
	Update(context.Context, *Debt) error
	DeleteByUserIdAndDebtId(context.Context, string, uint) error
	CountDebtDetailByDebtId(context.Context, uint) (int64, error)
	FindByID(context.Context, int64) (*Debt, error)
	FindByUserIdAndDebtId(context.Context, string, uint) (*Debt, error)
	ListByUserId(context.Context, string, *DebtListQuery) ([]*Debt, int64, error)
	ListByHello(context.Context, string) ([]*Debt, error)
	ListAll(context.Context) ([]*Debt, error)
}

type DebtUsecase struct {
	repo DebtRepo
	log  *log.Helper
}

func NewDebtUsecase(repo DebtRepo, logger log.Logger) *DebtUsecase {
	return &DebtUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *DebtUsecase) CreateDebt(ctx context.Context, g *Debt) (uint, error) {
	uc.log.WithContext(ctx).Infof("CreateDebt: %v", g.Id)
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}

	g.UserId = userId
	idDb, err := uc.repo.Save(ctx, g)
	return idDb, err
}
func (uc *DebtUsecase) Edit(ctx context.Context, g *Debt) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	debtDb, err := uc.repo.FindByUserIdAndDebtId(ctx, userId, uint(g.Id))
	if err != nil {
		return 0, err
	}
	g.Id = debtDb.Id
	g.UserId = userId
	err2 := uc.repo.Update(ctx, g)
	if err2 != nil {
		return 0, err
	}
	return uint(debtDb.Id), nil
}

func (uc *DebtUsecase) Delete(ctx context.Context, debtID uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}

	// Ownership check: ensure current user can only delete own debt.
	_, err = uc.repo.FindByUserIdAndDebtId(ctx, userId, debtID)
	if err != nil {
		return err
	}

	detailCount, err := uc.repo.CountDebtDetailByDebtId(ctx, debtID)
	if err != nil {
		return err
	}
	if detailCount > 0 {
		return errors.New("存在明细，禁止删除")
	}

	return uc.repo.DeleteByUserIdAndDebtId(ctx, userId, debtID)
}

func (uc *DebtUsecase) GetDebt(ctx context.Context, debtID uint) (*Debt, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return uc.repo.FindByUserIdAndDebtId(ctx, userId, debtID)
}

func (uc *DebtUsecase) ListDebt(ctx context.Context, query *DebtListQuery) ([]*Debt, int64, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, 0, err
	}
	return uc.repo.ListByUserId(ctx, userId, query)
}
