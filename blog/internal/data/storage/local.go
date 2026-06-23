package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"blog/internal/conf"

	"github.com/google/uuid"
)

type localStorage struct {
	uploadDir string
	baseURL   string
}

func newLocal(cfg *conf.Local) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("local config is nil")
	}
	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	// Ensure directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("local: create upload dir: %w", err)
	}
	abs, _ := filepath.Abs(uploadDir)
	return &localStorage{
		uploadDir: abs,
		baseURL:   "/files",
	}, nil
}

func (s *localStorage) Upload(ctx context.Context, fileName string, _ int64, _ string, reader io.Reader) (string, error) {
	ext := path.Ext(fileName)
	objectKey := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01/02"), uuid.New().String(), ext)
	fullPath := filepath.Join(s.uploadDir, objectKey)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("local: mkdir: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("local: create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("local: write file: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, objectKey), nil
}

func (s *localStorage) Delete(ctx context.Context, key string) error {
	// key is the URL path like "/files/2025/01/01/uuid.ext"
	// Strip the baseURL prefix to get the relative path
	relPath := key
	if len(key) > len(s.baseURL) && key[:len(s.baseURL)] == s.baseURL {
		relPath = key[len(s.baseURL)+1:]
	}
	fullPath := filepath.Join(s.uploadDir, relPath)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("local: delete: %w", err)
	}
	return nil
}
