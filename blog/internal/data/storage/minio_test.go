package storage

import "testing"

func TestMinioObjectKey(t *testing.T) {
	s := &minioStorage{
		bucket: "blog-files",
		prefix: "uploads",
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "presigned URL path-style",
			key:  "http://127.0.0.1:9000/blog-files/uploads/a.jpg?X-Amz-Signature=xyz",
			want: "uploads/a.jpg",
		},
		{
			name: "presigned URL virtual-host style",
			key:  "http://blog-files.127.0.0.1:9000/uploads/a.jpg?X-Amz-Signature=xyz",
			want: "uploads/a.jpg",
		},
		{
			name: "presigned URL https",
			key:  "https://minio.example.com/blog-files/uploads/a.jpg",
			want: "uploads/a.jpg",
		},
		{
			name: "bucket/key form",
			key:  "blog-files/uploads/a.jpg",
			want: "uploads/a.jpg",
		},
		{
			name: "bare object key",
			key:  "uploads/a.jpg",
			want: "uploads/a.jpg",
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
