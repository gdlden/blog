package storage

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestRustfsObjectKey(t *testing.T) {
	s := &rustfsStorage{
		bucket:   "blog-files",
		prefix:   "uploads",
		endpoint: "https://file.hukss.cn",
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "full URL returned by Upload",
			key:  "https://file.hukss.cn/blog-files/uploads/abc.jpg",
			want: "uploads/abc.jpg",
		},
		{
			name: "endpoint with trailing slash",
			key:  "https://file.hukss.cn//blog-files/uploads/abc.jpg",
			want: "uploads/abc.jpg",
		},
		{
			name: "bucket/key form",
			key:  "blog-files/uploads/abc.jpg",
			want: "uploads/abc.jpg",
		},
		{
			name: "bare object key",
			key:  "uploads/abc.jpg",
			want: "uploads/abc.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.objectKey(tt.key); got != tt.want {
				t.Errorf("objectKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestRustfsPresignedGetURL(t *testing.T) {
	cfgAws := aws.Config{
		Region: "default",
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: "https://file.hukss.cn"}, nil
		}),
		Credentials: credentials.NewStaticCredentialsProvider("test-ak", "test-sk", ""),
	}
	client := s3.NewFromConfig(cfgAws, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	s := &rustfsStorage{
		client:   client,
		bucket:   "blog-files",
		prefix:   "uploads",
		endpoint: "https://file.hukss.cn",
	}

	// 传入 Upload 返回的完整 URL，应提取出 object key 并生成签名 URL
	u, err := s.PresignedGetURL(context.Background(), "https://file.hukss.cn/blog-files/uploads/abc.jpg", 5*time.Minute)
	if err != nil {
		t.Fatalf("PresignedGetURL: %v", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}
	if parsed.Path != "/blog-files/uploads/abc.jpg" {
		t.Errorf("path = %q, want /blog-files/uploads/abc.jpg", parsed.Path)
	}
	q := parsed.Query()
	if q.Get("X-Amz-Signature") == "" {
		t.Error("missing X-Amz-Signature query parameter")
	}
	if q.Get("X-Amz-Expires") != "300" {
		t.Errorf("X-Amz-Expires = %q, want 300", q.Get("X-Amz-Expires"))
	}
}
