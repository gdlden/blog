package service

import (
	"context"
	"strconv"

	pb "blog/api/post/v1"
	"blog/internal/biz"
)

type PostService struct {
	pb.UnimplementedPostServer
	pu *biz.PostUsecase
}

func NewPostService(pu *biz.PostUsecase) *PostService {
	return &PostService{pu: pu}
}

func (s *PostService) CreatePost(ctx context.Context, req *pb.AddPostRequest) (*pb.AddPostReply, error) {
	s.pu.CreatePost(ctx, &biz.Post{
		Title:   req.Title,
		Content: req.Content,
	})
	return &pb.AddPostReply{}, nil
}
func (s *PostService) GetPostPage(ctx context.Context, req *pb.PostPageRequest) (*pb.PostPageReply, error) {
	current, _ := strconv.Atoi(req.Current)
	size, _ := strconv.Atoi(req.Size)

	posts, total, _ := s.pu.GetPostPage(ctx, &biz.PostPageRequest{
		Current: current,
		Size:    size,
	})
	if posts != nil {

		var postList []*pb.PostEntity
		for _, post := range posts {
			postList = append(postList, &pb.PostEntity{
				Id:      string(post.Id),
				Title:   post.Title,
				Content: post.Content,
			})
		}
		return &pb.PostPageReply{
			Data:  postList,
			Total: string(total),
		}, nil
	}
	return &pb.PostPageReply{}, nil
}

func (s *PostService) GetPostById(ctx context.Context, req *pb.GetPostByIdRequest) (*pb.GetPostByIdReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	post, _ := s.pu.GetPostById(ctx, id)
	if post != nil {
		return &pb.GetPostByIdReply{
			Id:      string(post.Id),
			Title:   post.Title,
			Content: post.Content,
		}, nil
	}
	return &pb.GetPostByIdReply{}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	post, _ := s.pu.UpdatePost(ctx, id, &biz.Post{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
	})
	if post != nil {
		return &pb.UpdatePostReply{
			Id:      post.Id,
			Title:   post.Title,
			Content: post.Content,
		}, nil
	}
	return &pb.UpdatePostReply{}, nil
}

func (s *PostService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	err := s.pu.DeletePost(ctx, id)
	return &pb.DeletePostReply{Success: err == nil}, nil
}
