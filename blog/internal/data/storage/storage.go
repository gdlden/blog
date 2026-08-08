package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"blog/internal/conf"
)

// Storage defines the interface for file storage backends.
type Storage interface {
	// Upload saves a file and returns its accessible URL or path.
	Upload(ctx context.Context, fileName string, fileSize int64, contentType string, reader io.Reader) (url string, err error)

	// Delete removes a file by its key (file path or object name).
	Delete(ctx context.Context, key string) error

	// GetReader returns a reader for downloading a file by its key.
	GetReader(ctx context.Context, key string) (io.ReadCloser, error)

	// PresignedGetURL returns a short-lived signed URL for direct download,
	// or an error if the backend does not support presigned URLs (e.g. local).
	PresignedGetURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// New creates a Storage backend based on config.
func New(c *conf.File) (Storage, error) {
	if c == nil {
		return nil, fmt.Errorf("file config is nil")
	}
	switch c.StorageType {
	case "minio":
		return newMinio(c.Minio)
	case "rustfs":
		return newRustfs(c.Rustfs)
	case "local":
		return newLocal(c.Local)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", c.StorageType)
	}
}
