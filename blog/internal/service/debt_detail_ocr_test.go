package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"blog/internal/biz"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeVisionTextRecognizer struct {
	image   string
	prompt  string
	rawText string
	err     error
}

func (r *fakeVisionTextRecognizer) RecognizeText(ctx context.Context, image string, prompt string) (string, error) {
	r.image = image
	r.prompt = prompt
	return r.rawText, r.err
}

type noopDebtDetailRepo struct{}

func (noopDebtDetailRepo) SaveDb(context.Context, *biz.DebtDetail) (string, error) {
	return "", nil
}

func (noopDebtDetailRepo) GetByUserIdAndID(context.Context, string, uint) (*biz.DebtDetail, error) {
	return nil, nil
}

func (noopDebtDetailRepo) ListByUserIdAndDebtId(context.Context, string, uint) ([]*biz.DebtDetail, error) {
	return nil, nil
}

func (noopDebtDetailRepo) EditDb(context.Context, *biz.DebtDetail) error {
	return nil
}

func (noopDebtDetailRepo) DeleteDb(context.Context, string, uint) error {
	return nil
}

func newDebtDetailOCRTestService(recognizer VisionTextRecognizer) *DebtDetailService {
	return NewDebtDetailServiceWithRecognizer(biz.NewDeptUseCase(noopDebtDetailRepo{}, log.DefaultLogger), recognizer)
}

type memoryMultipartFile struct {
	*bytes.Reader
}

func newMemoryMultipartFile(data []byte) *memoryMultipartFile {
	return &memoryMultipartFile{Reader: bytes.NewReader(data)}
}

func (f *memoryMultipartFile) Close() error {
	return nil
}

type trackingMultipartFile struct {
	*memoryMultipartFile
	closeCount int
}

func newTrackingMultipartFile(data []byte) *trackingMultipartFile {
	return &trackingMultipartFile{memoryMultipartFile: newMemoryMultipartFile(data)}
}

func (f *trackingMultipartFile) Close() error {
	f.closeCount++
	return nil
}

func newTestImageHeader(filename, contentType string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
		Size: size,
	}
}

func testJPEGBytes() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
}

func assertDebtDetailOCRBadRequest(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	kerr := kerrors.FromError(err)
	require.NotNil(t, kerr)
	assert.Equal(t, int32(http.StatusBadRequest), kerr.Code)
	assert.Equal(t, "DEBT_DETAIL_OCR_BAD_REQUEST", kerr.Reason)
}

func TestRecognizeDebtDetailOCR_ReturnsParsedItems(t *testing.T) {
	recognizer := &fakeVisionTextRecognizer{
		rawText: "第1期 本金¥100.00 利息¥1.20 入账日03-15",
	}
	service := newDebtDetailOCRTestService(recognizer)
	image := testJPEGBytes()

	reply, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
		DebtId: "9",
		Image:  newMemoryMultipartFile(image),
		Header: newTestImageHeader("debt.jpg", "image/jpeg", int64(len(image))),
		Year:   2026,
	})

	require.NoError(t, err)
	require.NotNil(t, reply)
	require.Len(t, reply.Items, 1)
	assert.Equal(t, "9", reply.Items[0].DebtId)
	assert.Equal(t, "2026-03-15 00:00:00", reply.Items[0].PostingDate)
	assert.Contains(t, recognizer.prompt, "待入账")
	assert.Contains(t, recognizer.image, "data:image/jpeg;base64,")
}

func TestRecognizeDebtDetailOCR_ParseFailureReturnsRawText(t *testing.T) {
	const rawText = "只识别到一些文字"
	recognizer := &fakeVisionTextRecognizer{rawText: rawText}
	service := newDebtDetailOCRTestService(recognizer)
	image := testJPEGBytes()

	reply, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
		DebtId: "9",
		Image:  newMemoryMultipartFile(image),
		Header: newTestImageHeader("debt.jpg", "image/jpeg", int64(len(image))),
		Year:   2026,
	})

	require.NoError(t, err)
	require.NotNil(t, reply)
	assert.Equal(t, rawText, reply.RawText)
	assert.Empty(t, reply.Items)
}

func TestRecognizeDebtDetailOCR_InvalidYearReturnsBadRequest(t *testing.T) {
	image := testJPEGBytes()
	service := newDebtDetailOCRTestService(&fakeVisionTextRecognizer{})

	_, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
		DebtId: "9",
		Image:  newMemoryMultipartFile(image),
		Header: newTestImageHeader("debt.jpg", "image/jpeg", int64(len(image))),
		Year:   1999,
	})

	assertDebtDetailOCRBadRequest(t, err)
}

func TestRecognizeDebtDetailOCR_NonImageReturnsBadRequest(t *testing.T) {
	data := []byte("this is not image data")
	service := newDebtDetailOCRTestService(&fakeVisionTextRecognizer{})

	_, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
		DebtId: "9",
		Image:  newMemoryMultipartFile(data),
		Header: newTestImageHeader("debt.txt", "text/plain", int64(len(data))),
		Year:   2026,
	})

	assertDebtDetailOCRBadRequest(t, err)
}

func TestRecognizeDebtDetailOCR_NilServiceReturnsErrorWithoutPanic(t *testing.T) {
	image := testJPEGBytes()
	var service *DebtDetailService

	require.NotPanics(t, func() {
		_, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
			DebtId: "9",
			Image:  newMemoryMultipartFile(image),
			Header: newTestImageHeader("debt.jpg", "image/jpeg", int64(len(image))),
			Year:   2026,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recognizer unavailable")
	})
}

func TestBuildOCRImageDataURI_RejectsOversizedImageHeader(t *testing.T) {
	image := testJPEGBytes()

	_, err := buildOCRImageDataURI(
		newMemoryMultipartFile(image),
		newTestImageHeader("debt.jpg", "image/jpeg", debtDetailOCRMaxImageBytes+1),
	)

	assertDebtDetailOCRBadRequest(t, err)
}

func TestBuildOCRImageDataURI_RejectsOversizedImageBody(t *testing.T) {
	image := append(testJPEGBytes(), bytes.Repeat([]byte{0}, debtDetailOCRMaxImageBytes+1)...)

	_, err := buildOCRImageDataURI(
		newMemoryMultipartFile(image),
		newTestImageHeader("debt.jpg", "image/jpeg", int64(len(testJPEGBytes()))),
	)

	assertDebtDetailOCRBadRequest(t, err)
}

func TestCleanupDebtDetailOCRRequest_ClosesImageAndNilsRequestImage(t *testing.T) {
	image := newTrackingMultipartFile(testJPEGBytes())
	req := &DebtDetailOCRRequest{Image: image}

	cleanupDebtDetailOCRRequest(nil, req)

	assert.Equal(t, 1, image.closeCount)
	assert.Nil(t, req.Image)
}

func TestParseDebtDetailOCRHTTPRequest_ReturnsMultipartFields(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("debtId", " 9 "))
	require.NoError(t, writer.WriteField("year", "2026"))
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="debt.jpg"`)
	partHeader.Set("Content-Type", "image/jpeg")
	part, err := writer.CreatePart(partHeader)
	require.NoError(t, err)
	_, err = part.Write(testJPEGBytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/debt-detail/ocr", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	req, err := parseDebtDetailOCRHTTPRequest(request)

	require.NoError(t, err)
	require.NotNil(t, req)
	defer req.Image.Close()
	assert.Equal(t, "9", req.DebtId)
	assert.Equal(t, 2026, req.Year)
	require.NotNil(t, req.Header)
	assert.Equal(t, "debt.jpg", req.Header.Filename)
	assert.Equal(t, "image/jpeg", req.Header.Header.Get("Content-Type"))
	data, err := io.ReadAll(req.Image)
	require.NoError(t, err)
	assert.Equal(t, testJPEGBytes(), data)
}

func TestParseDebtDetailOCRHTTPRequest_InvalidYearReturnsBadRequest(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("debtId", "9"))
	require.NoError(t, writer.WriteField("year", "bad-year"))
	part, err := writer.CreateFormFile("file", "debt.jpg")
	require.NoError(t, err)
	_, err = part.Write(testJPEGBytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/debt-detail/ocr", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	req, err := parseDebtDetailOCRHTTPRequest(request)

	assert.Nil(t, req)
	assertDebtDetailOCRBadRequest(t, err)
}

func TestRecognizeDebtDetailOCR_RecognizerErrorReturnsError(t *testing.T) {
	image := testJPEGBytes()
	service := newDebtDetailOCRTestService(&fakeVisionTextRecognizer{err: errors.New("ocr unavailable")})

	_, err := service.RecognizeDebtDetailOCR(context.Background(), &DebtDetailOCRRequest{
		DebtId: "9",
		Image:  newMemoryMultipartFile(image),
		Header: newTestImageHeader("debt.jpg", "image/jpeg", int64(len(image))),
		Year:   2026,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "debt detail ocr failed")
}
