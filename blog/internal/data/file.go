package data

import (
	"context"
	"io"
	"path"

	"blog/internal/biz"
	"blog/internal/conf"
	"blog/internal/data/storage"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type fileRecord struct {
	gorm.Model
	FileName    string `gorm:"column:file_name;type:varchar(255);not null;comment:原始文件名"`
	FilePath    string `gorm:"column:file_path;type:varchar(512);not null;comment:存储路径/对象键"`
	FileType    string `gorm:"column:file_type;type:varchar(128);comment:MIME类型"`
	FileExt     string `gorm:"column:file_ext;type:varchar(32);comment:文件后缀"`
	FileUrl     string `gorm:"column:file_url;type:text;comment:访问URL"`
	FileSize    int64  `gorm:"column:file_size;default:0;comment:文件大小(字节)"`
	StorageType string `gorm:"column:storage_type;type:varchar(32);comment:存储后端类型"`
}

func (fileRecord) TableName() string {
	return "file_records"
}

type fileRepo struct {
	data *Data
	log  *log.Helper
}

func NewFileRepo(data *Data, logger log.Logger) biz.FileRepo {
	return &fileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *fileRepo) Save(ctx context.Context, record *biz.FileRecord) (uint, error) {
	entity := fileRecord{
		FileName:    record.FileName,
		FilePath:    record.FilePath,
		FileType:    record.FileType,
		FileExt:     record.FileExt,
		FileUrl:     record.FileUrl,
		FileSize:    record.FileSize,
		StorageType: record.StorageType,
	}
	tx := r.data.db.WithContext(ctx).Create(&entity)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return entity.ID, nil
}

func (r *fileRepo) GetById(ctx context.Context, id uint) (*biz.FileRecord, error) {
	var entity fileRecord
	tx := r.data.db.WithContext(ctx).First(&entity, id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapFileRecordToBiz(&entity), nil
}

func mapFileRecordToBiz(entity *fileRecord) *biz.FileRecord {
	if entity == nil {
		return nil
	}
	return &biz.FileRecord{
		Id:          entity.ID,
		FileName:    entity.FileName,
		FilePath:    entity.FilePath,
		FileType:    entity.FileType,
		FileExt:     entity.FileExt,
		FileUrl:     entity.FileUrl,
		FileSize:    entity.FileSize,
		StorageType: entity.StorageType,
		CreatedAt:   entity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   entity.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// NewFileStorageFromConfig creates a file storage backend from config and wraps it as biz.FileStorage.
func NewFileStorageFromConfig(c *conf.File, logger log.Logger) biz.FileStorage {
	if c == nil || c.StorageType == "" {
		log.NewHelper(logger).Warn("file storage config is nil or empty, using noop storage")
		return &noopStorage{}
	}
	s, err := storage.New(c)
	if err != nil {
		log.NewHelper(logger).Warnf("failed to create file storage (%s), using noop: %v", c.StorageType, err)
		return &noopStorage{}
	}
	return &storageAdapter{s: s}
}

// storageAdapter adapts storage.Storage to biz.FileStorage.
type storageAdapter struct {
	s storage.Storage
}

func (a *storageAdapter) Upload(ctx context.Context, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error) {
	return a.s.Upload(ctx, fileName, fileSize, contentType, reader)
}

func (a *storageAdapter) Delete(ctx context.Context, key string) error {
	return a.s.Delete(ctx, key)
}

// noopStorage is a fallback that does nothing.
type noopStorage struct{}

func (n *noopStorage) Upload(ctx context.Context, _ string, _ int64, _ string, _ io.Reader) (string, error) {
	return "", nil
}

func (n *noopStorage) Delete(_ context.Context, _ string) error {
	return nil
}

// getFileExt returns the file extension from a filename.
func getFileExt(fileName string) string {
	return path.Ext(fileName)
}
