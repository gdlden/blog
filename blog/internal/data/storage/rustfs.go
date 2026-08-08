package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"blog/internal/conf"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type rustfsStorage struct {
	client   *s3.Client
	bucket   string
	prefix   string
	endpoint string
}

func newRustfs(cfg *conf.Rustfs) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rustfs config is nil")
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("rustfs endpoint is required")
	}
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "default"
	}
	region := cfg.Region
	if region == "" {
		region = "default"
	}

	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{URL: endpoint}, nil
	})

	cfgAws := aws.Config{
		Region:           region,
		EndpointResolver: resolver,
		Credentials:      credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}

	client := s3.NewFromConfig(cfgAws, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Ensure bucket exists (best-effort)
	_, err := client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		// Ignore BucketAlreadyOwnedByYou / BucketAlreadyExists errors
	}

	return &rustfsStorage{
		client:   client,
		bucket:   bucket,
		prefix:   "uploads",
		endpoint: endpoint,
	}, nil
}

func (s *rustfsStorage) Upload(ctx context.Context, fileName string, _ int64, contentType string, reader io.Reader) (string, error) {
	ext := path.Ext(fileName)
	objectKey := fmt.Sprintf("%s/%s%s", s.prefix, uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("rustfs: upload: %w", err)
	}

	// Return the accessible URL
	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectKey)
	return url, nil
}

func (s *rustfsStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(key)),
	})
	if err != nil {
		return fmt.Errorf("rustfs: delete: %w", err)
	}
	return nil
}

// objectKey 从存储层返回的 URL/路径中提取 S3 object key。
// Upload 返回的是完整 URL（{endpoint}/{bucket}/{objectKey}），历史数据也可能
// 是 "bucket/objectKey" 或纯 objectKey 形式；直接拿完整 URL 当 key 会 404。
func (s *rustfsStorage) objectKey(key string) string {
	endpoint := strings.TrimRight(s.endpoint, "/")
	prefix := endpoint + "/"
	if len(key) > len(prefix) && strings.HasPrefix(key, prefix) {
		// 兼容 endpoint 配置尾斜杠产生的 "endpoint//bucket/..." 形式
		rest := strings.TrimPrefix(key, prefix)
		rest = strings.TrimLeft(rest, "/")
		if strings.HasPrefix(rest, s.bucket+"/") {
			rest = strings.TrimPrefix(rest, s.bucket+"/")
		}
		if rest != "" {
			return rest
		}
	}
	bucketPrefix := s.bucket + "/"
	if len(key) > len(bucketPrefix) && strings.HasPrefix(key, bucketPrefix) {
		return key[len(bucketPrefix):]
	}
	return key
}

func (s *rustfsStorage) GetReader(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(key)),
	})
	if err != nil {
		return nil, fmt.Errorf("rustfs: get reader: %w", err)
	}
	return output.Body, nil
}
