package service

import (
	"context"

	pb "blog/api/file/v1"
)

type FileService struct {
	pb.UnimplementedFileServer
}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) CreateFileUploadUrl(ctx context.Context, req *pb.CreateFileUploadUrlRequest) (*pb.CreateFileUploadUrlReply, error) {
	return &pb.CreateFileUploadUrlReply{}, nil
}
func (s *FileService) CreateFile(ctx context.Context, req *pb.FileUploadRequest) (*pb.FileUploadReply, error) {
	return &pb.FileUploadReply{}, nil
}
func (s *FileService) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error) {
	return &pb.GetFileReply{}, nil
}
