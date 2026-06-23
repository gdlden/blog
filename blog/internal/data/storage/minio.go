package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"blog/internal/conf"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
	bucket string
	prefix string
}

func newMinio(cfg *conf.Minio) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("minio config is nil")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSsl,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create client: %w", err)
	}
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "default"
	}
	// Ensure bucket exists
	ctx := context.Background()
	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		// Ignore ErrBucketAlreadyExists / ErrBucketAlreadyOwnedByYou
		exists, errExists := client.BucketExists(ctx, bucket)
		if errExists != nil || !exists {
			return nil, fmt.Errorf("minio: ensure bucket %q: %w", bucket, err)
		}
	}
	return &minioStorage{
		client: client,
		bucket: bucket,
		prefix: "uploads",
	}, nil
}

func (s *minioStorage) Upload(ctx context.Context, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error) {
	ext := path.Ext(fileName)
	objectKey := fmt.Sprintf("%s/%s%s", s.prefix, uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, fileSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("minio: upload: %w", err)
	}

	// Generate presigned URL for download (valid 24h)
	reqParams := make(map[string][]string)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, 24*time.Hour, reqParams)
	if err != nil {
		// Fallback: return the object path directly
		return fmt.Sprintf("%s/%s", s.bucket, objectKey), nil
	}
	return presignedURL.String(), nil
}

func (s *minioStorage) Delete(ctx context.Context, key string) error {
	// The key may be a full pre-signed URL; extract object key from it,
	// or it may be a direct "bucket/key" or "key" path.
	objectKey := key
	// Try to parse as URL path segment after bucket
	if len(key) > len(s.bucket)+1 && key[:len(s.bucket)] == s.bucket {
		objectKey = key[len(s.bucket)+1:]
	}
	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: delete: %w", err)
	}
	return nil
}
