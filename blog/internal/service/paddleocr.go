package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var paddleOCRRecTextPattern = regexp.MustCompile(`['"]rec_text['"]\s*:\s*\[([^\]]*)\]`)
var paddleOCRQuotedTextPattern = regexp.MustCompile(`['"]([^'"]+)['"]`)

type PaddleOCRTextRecognizer struct {
	command string
	args    []string
	tempDir string
}

func NewPaddleOCRTextRecognizer(command string) *PaddleOCRTextRecognizer {
	command = strings.TrimSpace(command)
	if command == "" {
		command = defaultPaddleOCRCommand
	}
	parts := strings.Fields(command)
	if len(parts) == 1 && filepath.Base(parts[0]) == defaultPaddleOCRCommand {
		parts = append(parts, "ocr", "-i")
	}
	return newPaddleOCRTextRecognizer(parts[0], parts[1:])
}

func newPaddleOCRTextRecognizer(command string, args []string) *PaddleOCRTextRecognizer {
	return &PaddleOCRTextRecognizer{
		command: command,
		args:    append([]string{}, args...),
	}
}

func (r *PaddleOCRTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
	if r == nil || strings.TrimSpace(r.command) == "" {
		return "", fmt.Errorf("paddleocr command unavailable")
	}
	imageData, ext, err := decodeImageDataURI(imageURL)
	if err != nil {
		return "", err
	}
	imagePath, cleanup, err := writePaddleOCRTempImage(r.tempDir, imageData, ext)
	if err != nil {
		return "", err
	}
	defer cleanup()

	args := append(append([]string{}, r.args...), imagePath)
	cmd := exec.CommandContext(ctx, r.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("paddleocr failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return normalizePaddleOCROutput(string(output)), nil
}

func normalizePaddleOCROutput(output string) string {
	output = strings.TrimSpace(output)
	match := paddleOCRRecTextPattern.FindStringSubmatch(output)
	if len(match) != 2 {
		return output
	}
	quoted := paddleOCRQuotedTextPattern.FindAllStringSubmatch(match[1], -1)
	if len(quoted) == 0 {
		return output
	}
	texts := make([]string, 0, len(quoted))
	for _, item := range quoted {
		if len(item) == 2 && strings.TrimSpace(item[1]) != "" {
			texts = append(texts, strings.TrimSpace(item[1]))
		}
	}
	if len(texts) == 0 {
		return output
	}
	return strings.Join(texts, " ")
}

func decodeImageDataURI(imageURL string) ([]byte, string, error) {
	const marker = ";base64,"
	if !strings.HasPrefix(imageURL, "data:image/") {
		return nil, "", fmt.Errorf("invalid paddleocr image data")
	}
	parts := strings.SplitN(imageURL, marker, 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid paddleocr image data")
	}
	mimeType := strings.TrimPrefix(parts[0], "data:")
	mediaParts := strings.SplitN(mimeType, "/", 2)
	if len(mediaParts) != 2 || mediaParts[1] == "" {
		return nil, "", fmt.Errorf("invalid paddleocr image data")
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("invalid paddleocr image data: %w", err)
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("invalid paddleocr image data")
	}
	ext := "." + strings.TrimSpace(mediaParts[1])
	if decodedExt, err := url.PathUnescape(ext); err == nil {
		ext = decodedExt
	}
	ext = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' {
			return r
		}
		return -1
	}, ext)
	if ext == "." || ext == "" {
		ext = ".img"
	}
	return data, ext, nil
}

func writePaddleOCRTempImage(tempDir string, data []byte, ext string) (string, func(), error) {
	file, err := os.CreateTemp(tempDir, "paddleocr-*"+ext)
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return filepath.Clean(path), cleanup, nil
}
