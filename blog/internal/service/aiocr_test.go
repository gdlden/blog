package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
