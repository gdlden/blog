package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

var (
// ErrUserNotFound is user not found.
// ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Greeter is a Greeter model.
type Price struct {
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
