package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingVisionTextRecognizer struct {
	result string
	err    error
	calls  int
}

func (r *recordingVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
	r.calls++
	return r.result, r.err
}

func TestNewVisionTextRecognizerFromEnv_DefaultsToArk(t *testing.T) {
	t.Setenv("OCR_PROVIDER", "")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*ArkVisionTextRecognizer)
	assert.True(t, ok)
}

func TestNewVisionTextRecognizerFromEnv_SelectsArk(t *testing.T) {
	t.Setenv("OCR_PROVIDER", "ark")

	recognizer := NewVisionTextRecognizerFromEnv()

	_, ok := recognizer.(*ArkVisionTextRecognizer)
	assert.True(t, ok)
}

func TestNewVisionTextRecognizerFromEnv_SelectsPaddle(t *testing.T) {
	t.Setenv("OCR_PROVIDER", "paddle")
	t.Setenv("PADDLE_OCR_COMMAND", "custom-paddleocr")

	recognizer := NewVisionTextRecognizerFromEnv()

	paddle, ok := recognizer.(*PaddleOCRTextRecognizer)
	require.True(t, ok)
	assert.Equal(t, "custom-paddleocr", paddle.command)
}

func TestFallbackVisionTextRecognizer_ReturnsPrimaryResult(t *testing.T) {
	primary := &recordingVisionTextRecognizer{result: "local text"}
	secondary := &recordingVisionTextRecognizer{result: "ark text"}
	recognizer := NewFallbackVisionTextRecognizer(primary, secondary)

	result, err := recognizer.RecognizeText(context.Background(), "image", "prompt")

	require.NoError(t, err)
	assert.Equal(t, "local text", result)
	assert.Equal(t, 1, primary.calls)
	assert.Equal(t, 0, secondary.calls)
}

func TestFallbackVisionTextRecognizer_UsesSecondaryWhenPrimaryFails(t *testing.T) {
	primary := &recordingVisionTextRecognizer{err: errors.New("signal: killed")}
	secondary := &recordingVisionTextRecognizer{result: "ark text"}
	recognizer := NewFallbackVisionTextRecognizer(primary, secondary)

	result, err := recognizer.RecognizeText(context.Background(), "image", "prompt")

	require.NoError(t, err)
	assert.Equal(t, "ark text", result)
	assert.Equal(t, 1, primary.calls)
	assert.Equal(t, 1, secondary.calls)
}
