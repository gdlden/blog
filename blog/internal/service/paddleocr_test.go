package service

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaddleOCRTextRecognizer_RecognizeTextInvokesCommandWithTempImage(t *testing.T) {
	recognizer := newPaddleOCRTextRecognizer(os.Args[0], []string{
		"-test.run=TestPaddleOCRCommandHelper",
		"--",
		"success",
	})
	recognizer.tempDir = t.TempDir()

	rawText, err := recognizer.RecognizeText(
		context.Background(),
		"data:image/png;base64,iVBORw0KGgo=",
		"ignored prompt",
	)

	require.NoError(t, err)
	assert.Contains(t, rawText, "第1期 本金: 1000.00 利息: 12.34 入账日: 03-25")
}

func TestPaddleOCRTextRecognizer_RecognizeTextNormalizesRecTextOutput(t *testing.T) {
	recognizer := newPaddleOCRTextRecognizer(os.Args[0], []string{
		"-test.run=TestPaddleOCRCommandHelper",
		"--",
		"rec-text",
	})
	recognizer.tempDir = t.TempDir()

	rawText, err := recognizer.RecognizeText(
		context.Background(),
		"data:image/png;base64,iVBORw0KGgo=",
		"ignored prompt",
	)

	require.NoError(t, err)
	assert.Equal(t, "第1期 本金 1000.00 利息 12.34 入账日 03-25", rawText)
}

func TestNewPaddleOCRTextRecognizer_DefaultCommandUsesCondaPython(t *testing.T) {
	t.Setenv(envPaddleOCRPython, "custom-python.exe")
	recognizer := NewPaddleOCRTextRecognizer("paddleocr")

	assert.Equal(t, "custom-python.exe", recognizer.command)
	assert.Empty(t, recognizer.args, "sidecar mode has empty args")
	// sidecar mode uses the embedded server script (not in args)
	assert.True(t, recognizer.useSidecar, "paddleocr mode should use sidecar")
}

func TestNewPaddleOCRTextRecognizer_FallsBackToCustomCommand(t *testing.T) {
	recognizer := NewPaddleOCRTextRecognizer("custom-ocr --arg")

	assert.Equal(t, "custom-ocr", recognizer.command)
	assert.Equal(t, []string{"--arg"}, recognizer.args)
}

func TestNewPaddleOCRTextRecognizer_DefaultsPythonPathWhenEnvUnset(t *testing.T) {
	t.Setenv(envPaddleOCRPython, "")
	recognizer := NewPaddleOCRTextRecognizer("paddleocr")

	assert.Equal(t, defaultPaddleOCRPython, recognizer.command)
}

func TestNewPaddleOCRTextRecognizer_SidecarMode(t *testing.T) {
	t.Setenv(envPaddleOCRPython, "custom-python.exe")
	recognizer := NewPaddleOCRTextRecognizer("paddleocr")

	assert.Equal(t, "custom-python.exe", recognizer.command)
	assert.Empty(t, recognizer.args)
	assert.True(t, recognizer.useSidecar)

	// Verify human-readable command default
	t.Setenv(envPaddleOCRPython, "")
	recognizer2 := NewPaddleOCRTextRecognizer("paddleocr")
	assert.Equal(t, defaultPaddleOCRPython, recognizer2.command)
	assert.True(t, recognizer2.useSidecar)
}

func TestPaddleOCRTextRecognizer_RecognizeTextReturnsCommandError(t *testing.T) {
	recognizer := newPaddleOCRTextRecognizer(os.Args[0], []string{
		"-test.run=TestPaddleOCRCommandHelper",
		"--",
		"fail",
	})
	recognizer.tempDir = t.TempDir()

	_, err := recognizer.RecognizeText(
		context.Background(),
		"data:image/png;base64,iVBORw0KGgo=",
		"ignored prompt",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "paddleocr failed")
	assert.Contains(t, err.Error(), "simulated paddleocr failure")
}

func TestPaddleOCRTextRecognizer_RecognizeTextRejectsNonDataURI(t *testing.T) {
	recognizer := NewPaddleOCRTextRecognizer("paddleocr")

	_, err := recognizer.RecognizeText(context.Background(), "https://example.com/image.png", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid paddleocr image data")
}

func TestPaddleOCRCommandHelper(t *testing.T) {
	separator := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == -1 || len(os.Args) <= separator+2 {
		return
	}
	mode := os.Args[separator+1]
	if mode != "success" && mode != "fail" && mode != "rec-text" {
		return
	}

	imagePath := os.Args[len(os.Args)-1]
	if _, err := os.Stat(imagePath); err != nil {
		os.Stderr.WriteString("missing temp image")
		os.Exit(2)
	}

	if mode == "fail" {
		os.Stderr.WriteString("simulated paddleocr failure")
		os.Exit(3)
	}
	if mode == "rec-text" {
		os.Stdout.WriteString("{'rec_text': ['第1期', '本金', '1000.00', '利息', '12.34', '入账日', '03-25']}")
		os.Exit(0)
	}

	os.Stdout.WriteString(strings.Join([]string{
		"ocr result:",
		"第1期 本金: 1000.00 利息: 12.34 入账日: 03-25",
	}, "\n"))
	os.Exit(0)
}
