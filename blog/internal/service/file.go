package service

import (
	"context"
	"fmt"
	"io"
	"path"
	"strconv"

	pb "blog/api/file/v1"
	"blog/internal/biz"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"
)

type FileService struct {
	pb.UnimplementedFileServer
	fc *biz.FileUsecase
}

func NewFileService(fc *biz.FileUsecase) *FileService {
	return &FileService{fc: fc}
}

// CreateFileUploadUrl returns a file ID as a placeholder upload URL.
// For MinIO backends a pre-signed URL would be returned; for others
// the client uses the raw upload endpoint at POST /file/upload/raw/v1.
func (s *FileService) CreateFileUploadUrl(ctx context.Context, req *pb.CreateFileUploadUrlRequest) (*pb.CreateFileUploadUrlReply, error) {
	fileID := uuid.New().String()
	// Return the file ID encoded in the URL; client will use it
	// with the raw upload endpoint or the CreateFile endpoint.
	return &pb.CreateFileUploadUrlReply{
		Url: fileID,
	}, nil
}

// CreateFile saves a file record. In the pre-signed URL flow, the client
// uploads to the storage backend directly and then calls this to confirm.
func (s *FileService) CreateFile(ctx context.Context, req *pb.FileUploadRequest) (*pb.FileUploadReply, error) {
	fileID := req.FileId
	if fileID == "" {
		return &pb.FileUploadReply{SaveFlag: false}, nil
	}
	// Placeholder: in a full pre-signed flow we'd look up metadata by fileId
	// and save the record. For now, return true to indicate acceptance.
	return &pb.FileUploadReply{SaveFlag: true}, nil
}

// GetFile returns a file record by its database ID.
func (s *FileService) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error) {
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	record, err := s.fc.Get(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return &pb.GetFileReply{
		Id:       strconv.FormatUint(uint64(record.Id), 10),
		FileName: record.FileName,
		FilePath: record.FilePath,
		FileType: record.FileType,
		FileUrl:  record.FileUrl,
		FileExt:  record.FileExt,
	}, nil
}

// HandleRawUpload processes a raw file upload.
func (s *FileService) HandleRawUpload(ctx context.Context, fileName string, fileSize int64, contentType string, reader io.Reader) (id uint, url string, err error) {
	fileExt := path.Ext(fileName)
	return s.fc.Upload(ctx, fileName, fileSize, contentType, fileExt, reader)
}

// HandleRawUploadHTTP is the Kratos HTTP handler for raw file upload via multipart form.
// Route: POST /file/upload/raw/v1
func (s *FileService) HandleRawUploadHTTP(ctx http.Context) error {
	req := ctx.Request()
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	fileName := header.Filename
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileSize := header.Size

	id, _, err := s.HandleRawUpload(ctx, fileName, fileSize, contentType, file)
	if err != nil {
		return err
	}
	return ctx.Result(200, map[string]any{
		"id":  fmt.Sprintf("%d", id),
		"url": fmt.Sprintf("/file/download/v1/%d", id),
	})
}

// HandleDownloadHTTP streams a file from the storage backend to the client.
// Route: GET /file/download/v1/{id}
func (s *FileService) HandleDownloadHTTP(ctx http.Context) error {
	http.SetOperation(ctx, "/file.v1.File/Download")
	idValues := ctx.Vars()["id"]
	if len(idValues) == 0 {
		return fmt.Errorf("missing file id")
	}
	idStr := idValues[0]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return err
	}
	record, err := s.fc.Get(ctx, uint(id))
	if err != nil {
		return err
	}
	reader, err := s.fc.GetReader(ctx, record.FileUrl)
	if err != nil {
		return err
	}
	defer reader.Close()

	w := ctx.Response()
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, record.FileName))
	if record.FileType != "" {
		w.Header().Set("Content-Type", record.FileType)
	}
	w.WriteHeader(200)
	_, err = io.Copy(w, reader)
	return err
}
