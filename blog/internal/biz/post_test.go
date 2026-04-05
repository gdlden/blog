package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockPostRepo is a manual mock implementation of PostRepo interface
type mockPostRepo struct {
	saveFunc       func(context.Context, *Post) (*Post, error)
	updateFunc     func(context.Context, *Post) (*Post, error)
	deleteFunc     func(context.Context, int64) error
	findByIDFunc   func(context.Context, int64) (*Post, error)
	findByPageFunc func(context.Context, *PostPageRequest) ([]*Post, int64, error)
	listByHelloFunc func(context.Context, string) ([]*Post, error)
	listAllFunc    func(context.Context) ([]*Post, error)
}

func (m *mockPostRepo) Save(ctx context.Context, p *Post) (*Post, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, p)
	}
	return p, nil
}

func (m *mockPostRepo) Update(ctx context.Context, p *Post) (*Post, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return p, nil
}

func (m *mockPostRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockPostRepo) FindByID(ctx context.Context, id int64) (*Post, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockPostRepo) ListByHello(ctx context.Context, hello string) ([]*Post, error) {
	if m.listByHelloFunc != nil {
		return m.listByHelloFunc(ctx, hello)
	}
	return nil, nil
}

func (m *mockPostRepo) ListAll(ctx context.Context) ([]*Post, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockPostRepo) FindByPage(ctx context.Context, req *PostPageRequest) ([]*Post, int64, error) {
	if m.findByPageFunc != nil {
		return m.findByPageFunc(ctx, req)
	}
	return nil, 0, nil
}

func TestPostUsecase_CreatePost(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name    string
		post    *Post
		mockFn  func(ctx context.Context, p *Post) (*Post, error)
		want    *Post
		wantErr bool
	}{
		{
			name: "create post successfully",
			post: &Post{
				Title:   "Test Title",
				Content: "Test Content",
			},
			mockFn: func(ctx context.Context, p *Post) (*Post, error) {
				p.Id = "1"
				p.CreatedAt = "2024-01-01T00:00:00Z"
				return p, nil
			},
			want: &Post{
				Id:        "1",
				Title:     "Test Title",
				Content:   "Test Content",
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "create post with error",
			post: &Post{
				Title: "Error Post",
			},
			mockFn: func(ctx context.Context, p *Post) (*Post, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				saveFunc: tt.mockFn,
			}
			uc := NewPostUsecase(mock, logger)
			ctx := context.Background()

			got, err := uc.CreatePost(ctx, tt.post)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Content, got.Content)
		})
	}
}

func TestPostUsecase_GetPostPage(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mockPosts := []*Post{
		{Id: "1", Title: "Post 1", Content: "Content 1"},
		{Id: "2", Title: "Post 2", Content: "Content 2"},
		{Id: "3", Title: "Post 3", Content: "Content 3"},
	}

	mock := &mockPostRepo{
		findByPageFunc: func(ctx context.Context, req *PostPageRequest) ([]*Post, int64, error) {
			return mockPosts, 5, nil
		},
	}

	uc := NewPostUsecase(mock, logger)
	ctx := context.Background()

	posts, total, err := uc.GetPostPage(ctx, &PostPageRequest{Current: 1, Size: 3})

	assert.NoError(t, err)
	assert.Len(t, posts, 3)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, "1", posts[0].Id)
	assert.Equal(t, "Post 1", posts[0].Title)
}

func TestPostUsecase_GetPostById(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name    string
		id      int64
		mockFn  func(ctx context.Context, id int64) (*Post, error)
		want    *Post
		wantErr bool
	}{
		{
			name: "get existing post",
			id:   1,
			mockFn: func(ctx context.Context, id int64) (*Post, error) {
				return &Post{
					Id:      "1",
					Title:   "Test Post",
					Content: "Test Content",
				}, nil
			},
			want: &Post{
				Id:      "1",
				Title:   "Test Post",
				Content: "Test Content",
			},
			wantErr: false,
		},
		{
			name: "get non-existent post",
			id:   999,
			mockFn: func(ctx context.Context, id int64) (*Post, error) {
				return nil, errors.New("post not found")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				findByIDFunc: tt.mockFn,
			}
			uc := NewPostUsecase(mock, logger)
			ctx := context.Background()

			got, err := uc.GetPostById(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Title, got.Title)
		})
	}
}

func TestPostUsecase_UpdatePost(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name    string
		id      int64
		post    *Post
		mockFn  func(ctx context.Context, p *Post) (*Post, error)
		want    *Post
		wantErr bool
	}{
		{
			name: "update post successfully",
			id:   1,
			post: &Post{
				Id:      "1",
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			mockFn: func(ctx context.Context, p *Post) (*Post, error) {
				p.UpdatedAt = "2024-01-02T00:00:00Z"
				return p, nil
			},
			want: &Post{
				Id:        "1",
				Title:     "Updated Title",
				Content:   "Updated Content",
				UpdatedAt: "2024-01-02T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "update post with error",
			id:   1,
			post: &Post{
				Id:    "1",
				Title: "Error Update",
			},
			mockFn: func(ctx context.Context, p *Post) (*Post, error) {
				return nil, errors.New("update failed")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				updateFunc: tt.mockFn,
			}
			uc := NewPostUsecase(mock, logger)
			ctx := context.Background()

			got, err := uc.UpdatePost(ctx, tt.id, tt.post)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Title, got.Title)
		})
	}
}

func TestPostUsecase_DeletePost(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name    string
		id      int64
		mockFn  func(ctx context.Context, id int64) error
		wantErr bool
	}{
		{
			name: "delete post successfully",
			id:   1,
			mockFn: func(ctx context.Context, id int64) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "delete post with error",
			id:   999,
			mockFn: func(ctx context.Context, id int64) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				deleteFunc: tt.mockFn,
			}
			uc := NewPostUsecase(mock, logger)
			ctx := context.Background()

			err := uc.DeletePost(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
