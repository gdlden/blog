package service

import (
	"context"
	"testing"

	pb "blog/api/ocr/v1"
	"blog/internal/conf"

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

func TestNewVisionTextRecognizerFromEnv_ReturnsKimi(t *testing.T) {
	t.Setenv("OCR_API_KEY", "ds-test-key")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "should return KimiVisionTextRecognizer")
}

func TestNewVisionTextRecognizerFromEnv_ReturnsKimiEvenWithoutAnyKey(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "should return KimiVisionTextRecognizer even without API key")
}

func TestNewVisionTextRecognizerFromEnv_UsesKimiAPIKeyFallback(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "kimi-fallback-key")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	kimi, ok := recognizer.(*KimiVisionTextRecognizer)
	require.True(t, ok, "should return KimiVisionTextRecognizer")
	assert.NotNil(t, kimi.client)
}

func TestNewVisionTextRecognizerFromEnv_UsesMoonshotAPIKeyFallback(t *testing.T) {
	t.Setenv("OCR_API_KEY", "")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "moonshot-fallback-key")

	recognizer := NewVisionTextRecognizerFromEnv()

	kimi, ok := recognizer.(*KimiVisionTextRecognizer)
	require.True(t, ok, "should return KimiVisionTextRecognizer")
	assert.NotNil(t, kimi.client)
}

func TestNewKimiVisionTextRecognizer_UsesEnvModel(t *testing.T) {
	t.Setenv("OCR_MODEL", "moonshot-custom-model")

	recognizer := NewKimiVisionTextRecognizer("test-key")

	assert.Equal(t, "moonshot-custom-model", recognizer.model)
}

func TestNewKimiVisionTextRecognizer_DefaultModel(t *testing.T) {
	t.Setenv("OCR_MODEL", "")

	recognizer := NewKimiVisionTextRecognizer("test-key")

	assert.Equal(t, defaultKimiOCRModel, recognizer.model)
}

func TestNewVisionTextRecognizerFromConfig_UsesConfigKeyAndModel(t *testing.T) {
	t.Setenv("OCR_MODEL", "")

	recognizer := NewVisionTextRecognizerFromConfig(&conf.OCR{ApiKey: "cfg-key", Model: "cfg-model"})

	kimi, ok := recognizer.(*KimiVisionTextRecognizer)
	require.True(t, ok, "should return KimiVisionTextRecognizer")
	assert.Equal(t, "cfg-model", kimi.model)
	assert.NotNil(t, kimi.client)
}

func TestNewVisionTextRecognizerFromConfig_PlaceholderFallsBackToEnv(t *testing.T) {
	t.Setenv("OCR_API_KEY", "env-key")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromConfig(&conf.OCR{ApiKey: "${OCR_API_KEY}"})

	kimi, ok := recognizer.(*KimiVisionTextRecognizer)
	require.True(t, ok, "should fall back to env when config value is an unresolved placeholder")
	assert.NotNil(t, kimi.client)
}

func TestNewVisionTextRecognizerFromConfig_NilConfigFallsBackToEnv(t *testing.T) {
	t.Setenv("OCR_API_KEY", "env-key")
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromConfig(nil)

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "should return KimiVisionTextRecognizer for nil config")
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
