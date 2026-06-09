package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

var (
// ErrUserNotFound is user not found.
// ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Price is a price model.
type Price struct {
	ID        uint
	Name      string
	Price     string
	PriceDate string
}

// GreeterRepo is a Greater repo.
type PriceRep interface {
	Save(context.Context, *Price) uint
	Update(context.Context, *Price) (*Price, error)
	FindByID(context.Context, int64) (*Price, error)
	ListByHello(context.Context, string) ([]*Price, error)
	ListAll(context.Context) ([]*Price, error)
	Delete(context.Context, int64) error
}

// GreeterUsecase is a Greeter usecase.
type PriceUscase struct {
	repo PriceRep
	log  *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewPriceUsecase(repo PriceRep, logger log.Logger) *PriceUscase {
	return &PriceUscase{repo: repo, log: log.NewHelper(logger)}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *PriceUscase) CreatePrice(ctx context.Context, g *Price) uint {
	// uc.log.WithContext(ctx).Infof("CreateGreeter: %v", )
	return uc.repo.Save(ctx, g)
}

// ListAll returns all price records.
func (uc *PriceUscase) ListAll(ctx context.Context) ([]*Price, error) {
	return uc.repo.ListAll(ctx)
}

// GetPrice returns a price record by ID.
func (uc *PriceUscase) GetPrice(ctx context.Context, id int64) (*Price, error) {
	return uc.repo.FindByID(ctx, id)
}

// UpdatePrice updates a price record.
func (uc *PriceUscase) UpdatePrice(ctx context.Context, p *Price) (*Price, error) {
	return uc.repo.Update(ctx, p)
}

// DeletePrice deletes a price record by ID.
func (uc *PriceUscase) DeletePrice(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
