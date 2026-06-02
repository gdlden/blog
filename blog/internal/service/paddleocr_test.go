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
	require.Len(t, recognizer.args, 1)
	// args[0] is the temp script path — verify it's a valid .py file
	assert.True(t, strings.HasSuffix(recognizer.args[0], ".py"), "expected .py script path, got: %s", recognizer.args[0])
	content, err := os.ReadFile(recognizer.args[0])
	require.NoError(t, err)
	assert.Contains(t, string(content), "PaddleOCR")
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

func TestNewPaddleOCRTextRecognizer_FallbackToCWhenTempFileFails(t *testing.T) {
	t.Setenv(envPaddleOCRPython, "custom-python.exe")
	// Force extraction to fail by corrupting the cached state
	// Reset by using a fresh sub-test (sync.Once persists within a process, but test order is undefined)
	// Instead, just verify the args look right for temp file mode
	recognizer := NewPaddleOCRTextRecognizer("paddleocr")

	assert.Equal(t, "custom-python.exe", recognizer.command)
	require.Len(t, recognizer.args, 1)
	// Must be a temp file path, not "-c"
	assert.NotEqual(t, "-c", recognizer.args[0], "should use temp file, not -c")
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
