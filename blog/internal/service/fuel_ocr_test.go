package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubFuelOCRRecognizer struct {
	image   string
	prompt  string
	rawText string
	err     error
}

func (r *stubFuelOCRRecognizer) RecognizeText(ctx context.Context, image string, prompt string) (string, error) {
	r.image = image
	r.prompt = prompt
	return r.rawText, r.err
}

type memoryFuelOCRFile struct {
	*bytes.Reader
}

func newMemoryFuelOCRFile(data []byte) *memoryFuelOCRFile {
	return &memoryFuelOCRFile{Reader: bytes.NewReader(data)}
}

func (f *memoryFuelOCRFile) Close() error {
	return nil
}

func newFuelOCRImageHeader(filename, contentType string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
		Size: size,
	}
}

func fuelOCRJPEGBytes() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xff, 0xd9,
	}
}

func newFuelOCRTestService(recognizer VisionTextRecognizer) *FuelService {
	return NewFuelServiceWithRecognizer(nil, recognizer)
}

func newFuelOCRRequest(attachType string, data []byte) *FuelOCRRequest {
	return &FuelOCRRequest{
		AttachType: attachType,
		Image:      newMemoryFuelOCRFile(data),
		Header:     newFuelOCRImageHeader("fuel.jpg", "image/jpeg", int64(len(data))),
	}
}

func assertFuelOCRBadRequest(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	kerr := kerrors.FromError(err)
	require.NotNil(t, kerr)
	assert.Equal(t, int32(http.StatusBadRequest), kerr.Code)
	assert.Equal(t, "FUEL_OCR_BAD_REQUEST", kerr.Reason)
}

func TestRecognizeFuelOCR_StationScreenReturnsAmountVolumeUnitPrice(t *testing.T) {
	recognizer := &stubFuelOCRRecognizer{
		rawText: "金额: 225.00\n油量: 30.50\n单价: 7.38",
	}
	service := newFuelOCRTestService(recognizer)

	reply, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("station_screen", fuelOCRJPEGBytes()))

	require.NoError(t, err)
	require.NotNil(t, reply)
	assert.Equal(t, "225.00", reply.Amount)
	assert.Equal(t, "30.50", reply.Volume)
	assert.Equal(t, "7.38", reply.UnitPrice)
	assert.Equal(t, "", reply.Odometer)
	assert.Contains(t, reply.RawText, "金额")
	assert.Contains(t, recognizer.prompt, "加油机")
	assert.Contains(t, recognizer.image, "data:image/jpeg;base64,")
}

func TestRecognizeFuelOCR_DashboardReturnsOdometerOnly(t *testing.T) {
	recognizer := &stubFuelOCRRecognizer{
		rawText: "总里程: 52134",
	}
	service := newFuelOCRTestService(recognizer)

	reply, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("dashboard", fuelOCRJPEGBytes()))

	require.NoError(t, err)
	require.NotNil(t, reply)
	assert.Equal(t, "52134", reply.Odometer)
	assert.Equal(t, "", reply.Amount)
	assert.Equal(t, "", reply.Volume)
	assert.Equal(t, "", reply.UnitPrice)
	assert.Contains(t, recognizer.prompt, "仪表盘")
}

func TestRecognizeFuelOCR_InvalidAttachTypeReturnsBadRequest(t *testing.T) {
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{})

	for _, attachType := range []string{"", "receipt", "environment", "other"} {
		_, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest(attachType, fuelOCRJPEGBytes()))
		assertFuelOCRBadRequest(t, err)
	}
}

func TestRecognizeFuelOCR_RecognizerErrorReturnsError(t *testing.T) {
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{err: errors.New("ocr unavailable")})

	_, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("station_screen", fuelOCRJPEGBytes()))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fuel ocr failed")
}

func TestRecognizeFuelOCR_UnparseableTextReturnsRawTextWithEmptyFields(t *testing.T) {
	const rawText = "完全无法解析的文字"
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{rawText: rawText})

	reply, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("station_screen", fuelOCRJPEGBytes()))

	require.NoError(t, err)
	require.NotNil(t, reply)
	assert.Equal(t, rawText, reply.RawText)
	assert.Equal(t, "", reply.Amount)
	assert.Equal(t, "", reply.Volume)
	assert.Equal(t, "", reply.UnitPrice)
	assert.Equal(t, "", reply.Odometer)
}

func TestRecognizeFuelOCR_MissingFieldsStayEmpty(t *testing.T) {
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{rawText: "金额: 225.00"})

	reply, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("station_screen", fuelOCRJPEGBytes()))

	require.NoError(t, err)
	assert.Equal(t, "225.00", reply.Amount)
	assert.Equal(t, "", reply.Volume)
	assert.Equal(t, "", reply.UnitPrice)
}

func TestRecognizeFuelOCR_NonImageReturnsBadRequest(t *testing.T) {
	data := []byte("this is not image data")
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{})

	_, err := service.RecognizeFuelOCR(context.Background(), &FuelOCRRequest{
		AttachType: "station_screen",
		Image:      newMemoryFuelOCRFile(data),
		Header:     newFuelOCRImageHeader("fuel.txt", "text/plain", int64(len(data))),
	})

	// 图片校验复用 buildOCRImageDataURI，其 reason 为 DEBT_DETAIL_OCR_BAD_REQUEST，只断言 400
	require.Error(t, err)
	kerr := kerrors.FromError(err)
	require.NotNil(t, kerr)
	assert.Equal(t, int32(http.StatusBadRequest), kerr.Code)
}

func TestRecognizeFuelOCR_MissingImageReturnsBadRequest(t *testing.T) {
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{})

	_, err := service.RecognizeFuelOCR(context.Background(), &FuelOCRRequest{AttachType: "station_screen"})

	assertFuelOCRBadRequest(t, err)
}

func TestRecognizeFuelOCR_NilServiceReturnsErrorWithoutPanic(t *testing.T) {
	var service *FuelService

	require.NotPanics(t, func() {
		_, err := service.RecognizeFuelOCR(context.Background(), newFuelOCRRequest("station_screen", fuelOCRJPEGBytes()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recognizer unavailable")
	})
}

func TestRecognizeFuelOCR_NilRequestReturnsBadRequest(t *testing.T) {
	service := newFuelOCRTestService(&stubFuelOCRRecognizer{})

	_, err := service.RecognizeFuelOCR(context.Background(), nil)

	assertFuelOCRBadRequest(t, err)
}
