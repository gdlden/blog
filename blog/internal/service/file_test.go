package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blog/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
)

type mockFileRepo struct {
	file *biz.FileRecord
}

func (m *mockFileRepo) Save(_ context.Context, record *biz.FileRecord) (uint, error) {
	return record.Id, nil
}

func (m *mockFileRepo) GetById(_ context.Context, id uint) (*biz.FileRecord, error) {
	if m.file == nil || m.file.Id != id {
		return nil, errors.New("file record not found")
	}
	return m.file, nil
}

func (m *mockFileRepo) GetByIdAndUserId(ctx context.Context, id uint, _ string) (*biz.FileRecord, error) {
	return m.GetById(ctx, id)
}

type mockFileStore struct {
	presignedURL string
	presignErr   error
	content      string
}

func (m *mockFileStore) Upload(context.Context, string, int64, string, io.Reader) (string, error) {
	return "", nil
}

func (m *mockFileStore) Delete(context.Context, string) error {
	return nil
}

func (m *mockFileStore) GetReader(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(m.content)), nil
}

func (m *mockFileStore) PresignedGetURL(context.Context, string, time.Duration) (string, error) {
	return m.presignedURL, m.presignErr
}

func newFileDownloadServer(repo biz.FileRepo, store biz.FileStorage) *httptest.Server {
	uc := biz.NewFileUsecase(repo, store, log.DefaultLogger)
	svc := NewFileService(uc)
	srv := khttp.NewServer()
	srv.Route("/").GET("/file/download/v1/{id}", svc.HandleDownloadHTTP)
	return httptest.NewServer(srv)
}

func TestHandleDownloadHTTP_RedirectsToPresignedURL(t *testing.T) {
	repo := &mockFileRepo{file: &biz.FileRecord{
		Id:       9,
		FileName: "a.jpg",
		FileType: "image/jpeg",
		FileUrl:  "https://file.hukss.cn/blog-files/uploads/a.jpg",
	}}
	presigned := "https://file.hukss.cn/blog-files/uploads/a.jpg?X-Amz-Signature=abc&X-Amz-Expires=300"
	store := &mockFileStore{presignedURL: presigned}

	ts := newFileDownloadServer(repo, store)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(ts.URL + "/file/download/v1/9")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, presigned, resp.Header.Get("Location"))
}

func TestHandleDownloadHTTP_FallsBackToStreamingWhenNoPresign(t *testing.T) {
	repo := &mockFileRepo{file: &biz.FileRecord{
		Id:       9,
		FileName: "a.jpg",
		FileType: "image/jpeg",
		FileUrl:  "/files/2026/01/01/a.jpg",
	}}
	// local 等后端不支持签名：回退为流式代理下载
	store := &mockFileStore{presignErr: errors.New("local storage does not support presigned URLs"), content: "fake-image-bytes"}

	ts := newFileDownloadServer(repo, store)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/file/download/v1/9")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "fake-image-bytes", string(body))
}

func TestHandleDownloadHTTP_NotFound(t *testing.T) {
	repo := &mockFileRepo{} // 无记录
	store := &mockFileStore{}

	ts := newFileDownloadServer(repo, store)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/file/download/v1/999")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusFound, resp.StatusCode)
}
