package storage

import "testing"

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
