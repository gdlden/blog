package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"time"

	"blog/internal/conf"

	"github.com/google/uuid"
)

type rustfsStorage struct {
	endpoint string
	apiKey   string
	bucket   string
	client   *http.Client
}

type rustfsUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		URL  string `json:"url"`
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"data"`
}

func newRustfs(cfg *conf.Rustfs) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rustfs config is nil")
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080"
	}
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "default"
	}
	return &rustfsStorage{
		endpoint: endpoint,
		apiKey:   cfg.ApiKey,
		bucket:   bucket,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *rustfsStorage) Upload(ctx context.Context, fileName string, _ int64, contentType string, reader io.Reader) (string, error) {
	// Read the entire file into memory for multipart upload
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("rustfs: read file: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("rustfs: create form file: %w", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		return "", fmt.Errorf("rustfs: write form file: %w", err)
	}

	// Add extra fields
	if err := writer.WriteField("bucket", s.bucket); err != nil {
		return "", fmt.Errorf("rustfs: write field bucket: %w", err)
	}

	// Generate a unique object key
	ext := path.Ext(fileName)
	objectKey := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	if err := writer.WriteField("key", objectKey); err != nil {
		return "", fmt.Errorf("rustfs: write field key: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("rustfs: close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/api/upload", body)
	if err != nil {
		return "", fmt.Errorf("rustfs: create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("rustfs: upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("rustfs: upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var uploadResp rustfsUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", fmt.Errorf("rustfs: decode response: %w", err)
	}

	if uploadResp.Code != 0 || uploadResp.Data == nil {
		return "", fmt.Errorf("rustfs: upload error (code %d): %s", uploadResp.Code, uploadResp.Message)
	}

	return uploadResp.Data.URL, nil
}

func (s *rustfsStorage) Delete(ctx context.Context, key string) error {
	payload, _ := json.Marshal(map[string]string{"key": key})
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.endpoint+"/api/delete", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("rustfs: create delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("rustfs: delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rustfs: delete failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
