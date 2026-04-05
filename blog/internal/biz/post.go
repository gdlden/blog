package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// var (
// 	// ErrUserNotFound is user not found.
// 	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
// )

type Post struct {
	Id        string
	CreatedAt string
	UpdatedAt string
	// 其他字段...
	Title   string
	Content string
}
type PostPageRequest struct {
	Current int
	Size    int
}

type PostRepo interface {
	Save(context.Context, *Post) (*Post, error)
	Update(context.Context, *Post) (*Post, error)
	Delete(context.Context, int64) error
	FindByID(context.Context, int64) (*Post, error)
	ListByHello(context.Context, string) ([]*Post, error)
	ListAll(context.Context) ([]*Post, error)
	FindByPage(context.Context, *PostPageRequest) ([]*Post, int64, error)
}

type PostUsecase struct {
	repo PostRepo
	log  *log.Helper
}

func NewPostUsecase(repo PostRepo, logger log.Logger) *PostUsecase {
	return &PostUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *PostUsecase) CreatePost(ctx context.Context, g *Post) (*Post, error) {
	uc.log.WithContext(ctx).Infof("CreatePost: %v", g.Title)
	return uc.repo.Save(ctx, g)
}
func (uc *PostUsecase) GetPostPage(ctx context.Context, req *PostPageRequest) ([]*Post, int64, error) {
	uc.log.WithContext(ctx).Infof("GetPostPage: %v", req)
	return uc.repo.FindByPage(ctx, req)
}
func (uc *PostUsecase) GetPostById(ctx context.Context, id int64) (*Post, error) {
	uc.log.WithContext(ctx).Infof("GetPostById: %v", id)
	return uc.repo.FindByID(ctx, id)
}
