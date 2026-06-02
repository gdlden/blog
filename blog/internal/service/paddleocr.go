package service

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// One-shot helper script (used by tests and as sidecar content)
//
//go:embed paddleocr_helper.py
var paddleOCRHelperScript string

// Persistent sidecar server script
//
//go:embed paddleocr_server.py
var paddleOCRServerScript string

const defaultPaddleOCRPython = `C:\Users\hukss\anaconda3\python.exe`
const envPaddleOCRPython = "PADDLE_OCR_PYTHON"
const envPaddleOCRDisableEmbed = "PADDLE_OCR_DISABLE_EMBED"

var (
	extractServerOnce   sync.Once
	cachedServerPath    string
	cachedServerPathErr error
	normalizePaddleOCR  = sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(`['"]rec_text['"]\s*:\s*\[([^\]]*)\]`)
	})
)

var paddleOCRQuotedTextPattern = regexp.MustCompile(`['"]([^'"]+)['"]`)

type PaddleOCRTextRecognizer struct {
	command string
	args    []string
	tempDir string

	// sidecar process (only used when useSidecar is true)
	useSidecar   bool
	sidecarOnce  sync.Once
	sidecarCmd   *exec.Cmd
	sidecarStdin io.WriteCloser
	sidecarOut   *bufio.Scanner
	sidecarErr   error
}

func NewPaddleOCRTextRecognizer(command string) *PaddleOCRTextRecognizer {
	command = strings.TrimSpace(command)
	if command == "" {
		command = defaultPaddleOCRCommand
	}
	parts := strings.Fields(command)
	if len(parts) == 1 && filepath.Base(parts[0]) == defaultPaddleOCRCommand {
		pythonPath := strings.TrimSpace(os.Getenv(envPaddleOCRPython))
		if pythonPath == "" {
			pythonPath = defaultPaddleOCRPython
		}
		r := newPaddleOCRTextRecognizer(pythonPath, []string{})
		r.useSidecar = true
		return r
	}
	return newPaddleOCRTextRecognizer(parts[0], parts[1:])
}

func newPaddleOCRTextRecognizer(command string, args []string) *PaddleOCRTextRecognizer {
	return &PaddleOCRTextRecognizer{
		command: command,
		args:    append([]string{}, args...),
	}
}

// startSidecar lazily starts the persistent Python sidecar process (called once).
func (r *PaddleOCRTextRecognizer) startSidecar() {
	r.sidecarOnce.Do(func() {
		scriptPath, err := extractSidecarScript()
		if err != nil {
			r.sidecarErr = fmt.Errorf("extract sidecar script: %w", err)
			return
		}
		cmd := exec.Command(r.command, scriptPath)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			r.sidecarErr = fmt.Errorf("sidecar stdin pipe: %w", err)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			r.sidecarErr = fmt.Errorf("sidecar stdout pipe: %w", err)
			return
		}
		// Forward stderr to Go's stderr for debugging
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			r.sidecarErr = fmt.Errorf("sidecar start: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		// Wait for ready signal (first line)
		if !scanner.Scan() {
			_ = cmd.Wait()
			r.sidecarErr = fmt.Errorf("sidecar no ready signal")
			return
		}
		var ready struct {
			Ready bool `json:"ready"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &ready); err != nil || !ready.Ready {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			r.sidecarErr = fmt.Errorf("sidecar invalid ready signal")
			return
		}

		r.sidecarCmd = cmd
		r.sidecarStdin = stdin
		r.sidecarOut = scanner
	})
}

// killSidecar stops the sidecar process.
func (r *PaddleOCRTextRecognizer) killSidecar() {
	if r.sidecarCmd != nil && r.sidecarCmd.Process != nil {
		_ = r.sidecarCmd.Process.Kill()
		_ = r.sidecarCmd.Wait()
	}
}

// ensureSidecar checks if the sidecar is running; restarts if dead.
func (r *PaddleOCRTextRecognizer) ensureSidecar() error {
	r.startSidecar()
	if r.sidecarErr != nil {
		return r.sidecarErr
	}
	// If process has exited, restart by resetting Once
	if r.sidecarCmd != nil && r.sidecarCmd.ProcessState != nil && r.sidecarCmd.ProcessState.Exited() {
		r.sidecarOnce = sync.Once{}
		r.sidecarErr = nil
		r.startSidecar()
		return r.sidecarErr
	}
	return nil
}

// sendViaSidecar sends an image path to the sidecar and returns recognized texts.
func (r *PaddleOCRTextRecognizer) sendViaSidecar(ctx context.Context, imagePath string) (string, error) {
	// Build JSON request
	req := map[string]string{"image_path": imagePath}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("sidecar marshal request: %w", err)
	}

	// Write to stdin
	if _, err := fmt.Fprintln(r.sidecarStdin, string(reqJSON)); err != nil {
		return "", fmt.Errorf("sidecar write request: %w", err)
	}

	// Read response with context-aware goroutine
	type ocrResult struct {
		line string
		err  error
	}
	ch := make(chan ocrResult, 1)
	go func() {
		if r.sidecarOut.Scan() {
			ch <- ocrResult{line: r.sidecarOut.Text(), err: nil}
		} else {
			ch <- ocrResult{err: r.sidecarOut.Err()}
		}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return "", fmt.Errorf("sidecar read response: %w", res.err)
		}
		var resp struct {
			Texts []string `json:"texts"`
			Error *string  `json:"error"`
		}
		if err := json.Unmarshal([]byte(res.line), &resp); err != nil {
			return "", fmt.Errorf("sidecar invalid JSON: %w", err)
		}
		if resp.Error != nil {
			return "", fmt.Errorf("sidecar error: %s", *resp.Error)
		}
		return strings.Join(resp.Texts, " "), nil

	case <-ctx.Done():
		// Context cancelled — OCR might still complete; prefer result over error
		select {
		case res := <-ch:
			if res.err != nil {
				return "", fmt.Errorf("sidecar read response: %w", res.err)
			}
			var resp struct {
				Texts []string `json:"texts"`
				Error *string  `json:"error"`
			}
			if err := json.Unmarshal([]byte(res.line), &resp); err != nil {
				return "", fmt.Errorf("sidecar invalid JSON: %w", err)
			}
			if resp.Error != nil {
				return "", fmt.Errorf("sidecar error: %s", *resp.Error)
			}
			return strings.Join(resp.Texts, " "), nil
		default:
			// OCR really didn't finish — kill sidecar to clean up stale pipe data
			r.killSidecar()
			return "", ctx.Err()
		}
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

	// Try sidecar first (only for PaddleOCR mode)
	if r.useSidecar {
		if err := r.ensureSidecar(); err == nil {
			result, err := r.sendViaSidecar(ctx, imagePath)
			if err == nil {
				return result, nil
			}
			// Sidecar failed — fall through to one-shot
		}
	}

	// Fallback: one-shot subprocess
	var cmd *exec.Cmd
	if r.useSidecar {
		// sidecar mode: use one-shot helper script
		helperPath, err := extractOneShotHelper()
		if err != nil {
			return "", fmt.Errorf("paddleocr fallback: %w", err)
		}
		cmd = exec.CommandContext(ctx, r.command, helperPath, imagePath)
	} else {
		// custom command mode: use user-provided args
		args := append(append([]string{}, r.args...), imagePath)
		cmd = exec.CommandContext(ctx, r.command, args...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("paddleocr failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return normalizePaddleOCROutput(string(output)), nil
}

func extractSidecarScript() (string, error) {
	extractServerOnce.Do(func() {
		if path := strings.TrimSpace(os.Getenv(envPaddleOCRDisableEmbed)); path != "" {
			cachedServerPath = path
			return
		}
		f, err := os.CreateTemp("", "paddleocr_server-*.py")
		if err != nil {
			cachedServerPathErr = err
			return
		}
		path := f.Name()
		if _, err := f.WriteString(paddleOCRServerScript); err != nil {
			_ = f.Close()
			_ = os.Remove(path)
			cachedServerPathErr = err
			return
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(path)
			cachedServerPathErr = err
			return
		}
		cachedServerPath = path
	})
	return cachedServerPath, cachedServerPathErr
}

var (
	extractHelperOnce   sync.Once
	cachedHelperPath    string
	cachedHelperPathErr error
)

// extractOneShotHelper extracts the one-shot helper script to a temp file (once, cached).
// Used as fallback when the sidecar is unavailable.
func extractOneShotHelper() (string, error) {
	extractHelperOnce.Do(func() {
		if path := strings.TrimSpace(os.Getenv(envPaddleOCRDisableEmbed)); path != "" {
			cachedHelperPath = path
			return
		}
		f, err := os.CreateTemp("", "paddleocr_helper-*.py")
		if err != nil {
			cachedHelperPathErr = err
			return
		}
		path := f.Name()
		if _, err := f.WriteString(paddleOCRHelperScript); err != nil {
			_ = f.Close()
			_ = os.Remove(path)
			cachedHelperPathErr = err
			return
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(path)
			cachedHelperPathErr = err
			return
		}
		cachedHelperPath = path
	})
	return cachedHelperPath, cachedHelperPathErr
}

// normalizePaddleOCROutput parses one-shot PaddleOCR output (kept for fallback and test compatibility).
func normalizePaddleOCROutput(output string) string {
	output = strings.TrimSpace(output)
	recTextPattern := normalizePaddleOCR()
	match := recTextPattern.FindStringSubmatch(output)
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
