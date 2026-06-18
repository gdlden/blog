package service

import (
	"context"
	"testing"

	pb "blog/api/ocr/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingVisionTextRecognizer is a test double that records calls.
type recordingVisionTextRecognizer struct {
	result string
	err    error
	calls  int
}

func (r *recordingVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
	r.calls++
	return r.result, r.err
}

func TestNewVisionTextRecognizerFromEnv_ReturnsDeepSeek(t *testing.T) {
	t.Setenv("OCR_API_KEY", "ds-test-key")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*DeepSeekVisionTextRecognizer)
	assert.True(t, ok, "should return DeepSeekVisionTextRecognizer")
}

func TestNewVisionTextRecognizerFromEnv_ReturnsDeepSeekEvenWithoutAnyKey(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*DeepSeekVisionTextRecognizer)
	assert.True(t, ok, "should return DeepSeekVisionTextRecognizer even without API key")
}

func TestNewVisionTextRecognizerFromEnv_UsesKimiAPIKeyFallback(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "kimi-fallback-key")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	ds, ok := recognizer.(*DeepSeekVisionTextRecognizer)
	require.True(t, ok, "should return DeepSeekVisionTextRecognizer")
	assert.NotNil(t, ds.client)
}

func TestNewVisionTextRecognizerFromEnv_UsesMoonshotAPIKeyFallback(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "moonshot-fallback-key")

	recognizer := NewVisionTextRecognizerFromEnv()

	ds, ok := recognizer.(*DeepSeekVisionTextRecognizer)
	require.True(t, ok, "should return DeepSeekVisionTextRecognizer")
	assert.NotNil(t, ds.client)
}

func TestNewDeepSeekVisionTextRecognizer_UsesEnvModel(t *testing.T) {
	t.Setenv("OCR_MODEL", "deepseek-custom-model")

	recognizer := NewDeepSeekVisionTextRecognizer("test-key")

	assert.Equal(t, "deepseek-custom-model", recognizer.model)
}

func TestNewDeepSeekVisionTextRecognizer_DefaultModel(t *testing.T) {
	t.Setenv("OCR_MODEL", "")

	recognizer := NewDeepSeekVisionTextRecognizer("test-key")

	assert.Equal(t, defaultDeepSeekOCRModel, recognizer.model)
}

func TestAiocrService_Ocr_DelegatesToRecognizer(t *testing.T) {
	rec := &recordingVisionTextRecognizer{result: "ocr result text"}
	service := NewAiocrServiceWithRecognizer(rec)

	reply, err := service.Ocr(context.Background(), &pb.OcrRequest{
		ImgBaseStr: "data:image/png;base64,abc123",
	})

	require.NoError(t, err)
	assert.Equal(t, "ocr result text", reply.Res)
	assert.Equal(t, 1, rec.calls)
}

func TestAiocrService_Ocr_ReturnsErrorFromRecognizer(t *testing.T) {
	rec := &recordingVisionTextRecognizer{err: assert.AnError}
	service := NewAiocrServiceWithRecognizer(rec)

	_, err := service.Ocr(context.Background(), &pb.OcrRequest{
		ImgBaseStr: "data:image/png;base64,abc123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ocr failed")
	assert.Equal(t, 1, rec.calls)
}
