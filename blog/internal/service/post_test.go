package service

import (
	"context"
	"errors"
	"testing"

	pb "blog/api/post/v1"
	"blog/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockPostRepo is a manual mock implementation of biz.PostRepo interface
type mockPostRepo struct {
	saveFunc       func(context.Context, *biz.Post) (*biz.Post, error)
	updateFunc     func(context.Context, *biz.Post) (*biz.Post, error)
	deleteFunc     func(context.Context, int64) error
	findByIDFunc   func(context.Context, int64) (*biz.Post, error)
	findByPageFunc func(context.Context, *biz.PostPageRequest) ([]*biz.Post, int64, error)
	listByHelloFunc func(context.Context, string) ([]*biz.Post, error)
	listAllFunc    func(context.Context) ([]*biz.Post, error)
}

func (m *mockPostRepo) Save(ctx context.Context, p *biz.Post) (*biz.Post, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, p)
	}
	return p, nil
}

func (m *mockPostRepo) Update(ctx context.Context, p *biz.Post) (*biz.Post, error) {
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

func (m *mockPostRepo) FindByID(ctx context.Context, id int64) (*biz.Post, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockPostRepo) ListByHello(ctx context.Context, hello string) ([]*biz.Post, error) {
	if m.listByHelloFunc != nil {
		return m.listByHelloFunc(ctx, hello)
	}
	return nil, nil
}

func (m *mockPostRepo) ListAll(ctx context.Context) ([]*biz.Post, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockPostRepo) FindByPage(ctx context.Context, req *biz.PostPageRequest) ([]*biz.Post, int64, error) {
	if m.findByPageFunc != nil {
		return m.findByPageFunc(ctx, req)
	}
	return nil, 0, nil
}

// createPostServiceWithMock creates a PostService with a mocked repository
func createPostServiceWithMock(mockRepo *mockPostRepo) *PostService {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	uc := biz.NewPostUsecase(mockRepo, logger)
	return NewPostService(uc)
}

func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.AddPostRequest
		mockFn  func(context.Context, *biz.Post) (*biz.Post, error)
		wantErr bool
	}{
		{
			name: "create post successfully",
			req: &pb.AddPostRequest{
				Title:   "Test Title",
				Content: "Test Content",
			},
			mockFn: func(ctx context.Context, p *biz.Post) (*biz.Post, error) {
				p.Id = "1"
				return p, nil
			},
			wantErr: false,
		},
		{
			name: "create post with error",
			req: &pb.AddPostRequest{
				Title:   "Error Post",
				Content: "Error Content",
			},
			mockFn: func(ctx context.Context, p *biz.Post) (*biz.Post, error) {
				return nil, errors.New("database error")
			},
			wantErr: false, // Service layer doesn't return error from usecase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				saveFunc: tt.mockFn,
			}
			svc := createPostServiceWithMock(mock)
			ctx := context.Background()

			reply, err := svc.CreatePost(ctx, tt.req)
			assert.NoError(t, err)
			assert.NotNil(t, reply)
		})
	}
}

func TestPostService_GetPostPage(t *testing.T) {
	mockPosts := []*biz.Post{
		{Id: "1", Title: "Post 1", Content: "Content 1"},
		{Id: "2", Title: "Post 2", Content: "Content 2"},
		{Id: "3", Title: "Post 3", Content: "Content 3"},
	}

	mock := &mockPostRepo{
		findByPageFunc: func(ctx context.Context, req *biz.PostPageRequest) ([]*biz.Post, int64, error) {
			return mockPosts, 5, nil
		},
	}

	svc := createPostServiceWithMock(mock)
	ctx := context.Background()

	req := &pb.PostPageRequest{
		Current: "1",
		Size:    "3",
	}

	reply, err := svc.GetPostPage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Len(t, reply.Data, 3)
	assert.Equal(t, "5", reply.Total)
	assert.Equal(t, "1", reply.Data[0].Id)
	assert.Equal(t, "Post 1", reply.Data[0].Title)
}

func TestPostService_GetPostById(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.GetPostByIdRequest
		mockFn    func(context.Context, int64) (*biz.Post, error)
		wantId    string
		wantTitle string
		wantErr   bool
	}{
		{
			name: "get existing post",
			req: &pb.GetPostByIdRequest{
				Id: "1",
			},
			mockFn: func(ctx context.Context, id int64) (*biz.Post, error) {
				return &biz.Post{
					Id:      "1",
					Title:   "Test Post",
					Content: "Test Content",
				}, nil
			},
			wantId:    "1",
			wantTitle: "Test Post",
			wantErr:   false,
		},
		{
			name: "get non-existent post",
			req: &pb.GetPostByIdRequest{
				Id: "999",
			},
			mockFn: func(ctx context.Context, id int64) (*biz.Post, error) {
				return nil, errors.New("post not found")
			},
			wantId:    "",
			wantTitle: "",
			wantErr:   false, // Service returns empty reply on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				findByIDFunc: tt.mockFn,
			}
			svc := createPostServiceWithMock(mock)
			ctx := context.Background()

			reply, err := svc.GetPostById(ctx, tt.req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantId, reply.Id)
			assert.Equal(t, tt.wantTitle, reply.Title)
		})
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.UpdatePostRequest
		mockFn    func(context.Context, *biz.Post) (*biz.Post, error)
		wantTitle string
		wantErr   bool
	}{
		{
			name: "update post successfully",
			req: &pb.UpdatePostRequest{
				Id:      "1",
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			mockFn: func(ctx context.Context, p *biz.Post) (*biz.Post, error) {
				return &biz.Post{
					Id:      "1",
					Title:   "Updated Title",
					Content: "Updated Content",
				}, nil
			},
			wantTitle: "Updated Title",
			wantErr:   false,
		},
		{
			name: "update post with error",
			req: &pb.UpdatePostRequest{
				Id:      "1",
				Title:   "Error Update",
				Content: "Error Content",
			},
			mockFn: func(ctx context.Context, p *biz.Post) (*biz.Post, error) {
				return nil, errors.New("update failed")
			},
			wantTitle: "",
			wantErr:   false, // Service returns empty reply on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				updateFunc: tt.mockFn,
			}
			svc := createPostServiceWithMock(mock)
			ctx := context.Background()

			reply, err := svc.UpdatePost(ctx, tt.req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTitle, reply.Title)
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.DeletePostRequest
		mockFn      func(context.Context, int64) error
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "delete post successfully",
			req: &pb.DeletePostRequest{
				Id: "1",
			},
			mockFn: func(ctx context.Context, id int64) error {
				return nil
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "delete post with error",
			req: &pb.DeletePostRequest{
				Id: "999",
			},
			mockFn: func(ctx context.Context, id int64) error {
				return errors.New("delete failed")
			},
			wantSuccess: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPostRepo{
				deleteFunc: tt.mockFn,
			}
			svc := createPostServiceWithMock(mock)
			ctx := context.Background()

			reply, err := svc.DeletePost(ctx, tt.req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSuccess, reply.Success)
		})
	}
}
