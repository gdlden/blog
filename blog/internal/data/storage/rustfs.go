package storage

import (
	"context"
	"fmt"
	"io"
	"path"

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
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("rustfs: delete: %w", err)
	}
	return nil
}

func (s *rustfsStorage) GetReader(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("rustfs: get reader: %w", err)
	}
	return output.Body, nil
}
