package data

import (
	"context"
	"strconv"

	"blog/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type postRepo struct {
	data *Data
	log  *log.Helper
}
type Post struct {
	gorm.Model
	Title   string
	Content string
}

// NewGreeterRepo .
func NewPostRepo(data *Data, logger log.Logger) biz.PostRepo {
	return &postRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *postRepo) Save(ctx context.Context, g *biz.Post) (*biz.Post, error) {
	var post = Post{
		Title:   g.Title,
		Content: g.Content,
	}
	res := r.data.db.Create(&post)
	if res.RowsAffected <= 0 {
		log.Info("插入失败")
	}
	g.Id = strconv.FormatUint(uint64(post.ID), 10)
	return g, nil
}

func (r *postRepo) Update(ctx context.Context, g *biz.Post) (*biz.Post, error) {
	id, _ := strconv.ParseUint(g.Id, 10, 64)
	err := r.data.db.Model(&Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":   g.Title,
		"content": g.Content,
	}).Error
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *postRepo) Delete(ctx context.Context, id int64) error {
	return r.data.db.Delete(&Post{}, id).Error
}

func (r *postRepo) FindByID(ctx context.Context, id int64) (*biz.Post, error) {
	var post Post
	err := r.data.db.Where("id = ?", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &biz.Post{
		Id:      strconv.FormatUint(uint64(post.ID), 10),
		Title:   post.Title,
		Content: post.Content,
	}, nil
}

func (r *postRepo) ListByHello(context.Context, string) ([]*biz.Post, error) {
	return nil, nil
}

func (r *postRepo) ListAll(context.Context) ([]*biz.Post, error) {
	return nil, nil
}
func (r *postRepo) FindByPage(ctx context.Context, req *biz.PostPageRequest) ([]*biz.Post, int64, error) {
	var posts []*biz.Post
	var count int64
	err := r.data.db.WithContext(ctx).Model(&Post{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.data.db.WithContext(ctx).Model(&Post{}).Offset(int((req.Current - 1) * req.Size)).Limit(int(req.Size)).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	return posts, count, nil
}
