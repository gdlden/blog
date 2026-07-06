package biz

import (
	"blog/internal/utils"
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// Price is a price model for grocery receipts.
type Price struct {
	ID          uint
	ProductName string
	Weight      decimal.Decimal
	UnitPrice   decimal.Decimal
	TotalPrice  decimal.Decimal
	PriceDate   string
	UserId      string
}

type PricePageRequest struct {
	Current int
	Size    int
}

// PriceRep is the price repository interface.
type PriceRep interface {
	Save(context.Context, *Price) (uint, error)
	Update(context.Context, *Price) error
	FindByUserIdAndID(context.Context, string, uint) (*Price, error)
	FindByPage(context.Context, string, *PricePageRequest) ([]*Price, int64, error)
	DeleteByUserIdAndID(context.Context, string, uint) error
}

// PriceUscase is the price use case.
type PriceUscase struct {
	repo PriceRep
	log  *log.Helper
}

// NewPriceUsecase creates a new PriceUscase.
func NewPriceUsecase(repo PriceRep, logger log.Logger) *PriceUscase {
	return &PriceUscase{repo: repo, log: log.NewHelper(logger)}
}

// GetPricePage returns paginated price records for current user.
func (uc *PriceUscase) GetPricePage(ctx context.Context, req *PricePageRequest) ([]*Price, int64, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, 0, err
	}
	prices, total, err := uc.repo.FindByPage(ctx, userId, req)
	if err != nil {
		return nil, 0, err
	}
	// Calculate total price for each record
	for _, p := range prices {
		p.TotalPrice = p.Weight.Mul(p.UnitPrice).Round(2)
	}
	return prices, total, nil
}

// CreatePrice creates a new price record.
func (uc *PriceUscase) CreatePrice(ctx context.Context, g *Price) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	g.UserId = userId
	return uc.repo.Save(ctx, g)
}

// GetPrice returns a price record by ID for current user.
func (uc *PriceUscase) GetPrice(ctx context.Context, id uint) (*Price, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	p, err := uc.repo.FindByUserIdAndID(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	p.TotalPrice = p.Weight.Mul(p.UnitPrice).Round(2)
	return p, nil
}

// UpdatePrice updates a price record for current user.
func (uc *PriceUscase) UpdatePrice(ctx context.Context, p *Price) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	// Verify ownership
	_, err = uc.repo.FindByUserIdAndID(ctx, userId, p.ID)
	if err != nil {
		return err
	}
	p.UserId = userId
	return uc.repo.Update(ctx, p)
}

// DeletePrice deletes a price record by ID for current user.
func (uc *PriceUscase) DeletePrice(ctx context.Context, id uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	return uc.repo.DeleteByUserIdAndID(ctx, userId, id)
}
