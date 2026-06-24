package biz

import (
	"context"
	"io"

	"github.com/go-kratos/kratos/v2/log"
)

type FileRecord struct {
	Id          uint
	FileName    string
	FilePath    string
	FileType    string
	FileExt     string
	FileUrl     string
	FileSize    int64
	StorageType string
	CreatedAt   string
	UpdatedAt   string
}

type FileRepo interface {
	Save(ctx context.Context, record *FileRecord) (uint, error)
	GetById(ctx context.Context, id uint) (*FileRecord, error)
}

type FileUsecase struct {
	repo  FileRepo
	store FileStorage
	log   *log.Helper
}

// FileStorage is the minimal interface the file usecase needs from storage.
// It matches the storage.Storage interface signature.
type FileStorage interface {
	Upload(ctx context.Context, fileName string, fileSize int64, contentType string, reader io.Reader) (url string, err error)
	Delete(ctx context.Context, key string) error
	GetReader(ctx context.Context, key string) (io.ReadCloser, error)
}

func NewFileUsecase(repo FileRepo, store FileStorage, logger log.Logger) *FileUsecase {
	return &FileUsecase{
		repo:  repo,
		store: store,
		log:   log.NewHelper(logger),
	}
}

// Upload uploads a file to the configured storage backend and saves a record to DB.
// Returns the file record ID and the accessible URL.
func (uc *FileUsecase) Upload(ctx context.Context, fileName string, fileSize int64, contentType string, fileExt string, reader io.Reader) (id uint, url string, err error) {
	url, err = uc.store.Upload(ctx, fileName, fileSize, contentType, reader)
	if err != nil {
		return 0, "", err
	}

	record := &FileRecord{
		FileName: fileName,
		FilePath: url,
		FileType: contentType,
		FileExt:  fileExt,
		FileUrl:  url,
		FileSize: fileSize,
	}

	savedId, err := uc.repo.Save(ctx, record)
	if err != nil {
		return 0, "", err
	}
	return savedId, url, nil
}

// Get returns a file record by ID.
func (uc *FileUsecase) Get(ctx context.Context, id uint) (*FileRecord, error) {
	return uc.repo.GetById(ctx, id)
}

// GetReader returns a reader for downloading a file by the stored URL.
func (uc *FileUsecase) GetReader(ctx context.Context, fileUrl string) (io.ReadCloser, error) {
	return uc.store.GetReader(ctx, fileUrl)
}
