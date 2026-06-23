package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type AppVersion struct {
	Id         uint
	Version    string
	Info       []string
	IosUrl     string
	AndroidUrl string
	IsActive   bool
	CreatedAt  string
	UpdatedAt  string
}

type AppVersionPageRequest struct {
	Current  int
	PageSize int
}

type AppVersionRepo interface {
	Save(ctx context.Context, av *AppVersion) (uint, error)
	Update(ctx context.Context, av *AppVersion) error
	Delete(ctx context.Context, id uint) error
	GetById(ctx context.Context, id uint) (*AppVersion, error)
	ListPage(ctx context.Context, req *AppVersionPageRequest) ([]*AppVersion, int64, error)
	GetActive(ctx context.Context) (*AppVersion, error)
	SetActive(ctx context.Context, id uint) error
	ClearAllActive(ctx context.Context) error
}

type AppVersionUsecase struct {
	repo AppVersionRepo
	log  *log.Helper
}

func NewAppVersionUsecase(repo AppVersionRepo, logger log.Logger) *AppVersionUsecase {
	return &AppVersionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *AppVersionUsecase) Create(ctx context.Context, av *AppVersion) (uint, error) {
	return uc.repo.Save(ctx, av)
}

func (uc *AppVersionUsecase) Update(ctx context.Context, av *AppVersion) error {
	return uc.repo.Update(ctx, av)
}

func (uc *AppVersionUsecase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *AppVersionUsecase) Get(ctx context.Context, id uint) (*AppVersion, error) {
	return uc.repo.GetById(ctx, id)
}

func (uc *AppVersionUsecase) ListPage(ctx context.Context, req *AppVersionPageRequest) ([]*AppVersion, int64, error) {
	return uc.repo.ListPage(ctx, req)
}

func (uc *AppVersionUsecase) GetActive(ctx context.Context) (*AppVersion, error) {
	return uc.repo.GetActive(ctx)
}
