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

func TestNewVisionTextRecognizerFromEnv_ReturnsKimi(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "kimi-test-key")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "should return KimiVisionTextRecognizer")
}

func TestNewVisionTextRecognizerFromEnv_ReturnsKimiEvenWithoutAnyKey(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "should return KimiVisionTextRecognizer even without API key")
}

func TestNewVisionTextRecognizerFromEnv_UsesMoonshotAPIKeyFallback(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "")
	t.Setenv("MOONSHOT_API_KEY", "moonshot-fallback-key")

	recognizer := NewVisionTextRecognizerFromEnv()

	kimi, ok := recognizer.(*KimiVisionTextRecognizer)
	require.True(t, ok, "should return KimiVisionTextRecognizer")
	// The client is constructed with the Moonshot key internally;
	// we verify the recognizer was created without panicking.
	assert.NotNil(t, kimi.client)
}

func TestNewKimiVisionTextRecognizer_UsesEnvModel(t *testing.T) {
	t.Setenv("KIMI_OCR_MODEL", "kimi-custom-model")

	recognizer := NewKimiVisionTextRecognizer("test-key")

	assert.Equal(t, "kimi-custom-model", recognizer.model)
}

func TestNewKimiVisionTextRecognizer_DefaultModel(t *testing.T) {
	t.Setenv("KIMI_OCR_MODEL", "")

	recognizer := NewKimiVisionTextRecognizer("test-key")

	assert.Equal(t, defaultKimiOCRModel, recognizer.model)
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
