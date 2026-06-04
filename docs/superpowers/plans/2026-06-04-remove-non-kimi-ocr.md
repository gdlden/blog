# Remove All Non-Kimi OCR Providers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all OCR providers (PaddleOCR, Ark/Doubao, Fallback) from the aiocr module, keeping only the Kimi (Moonshot) vision text recognizer.

**Architecture:** Delete `paddleocr.go` and `paddleocr_test.go` entirely. Strip `aiocr.go` down to: `VisionTextRecognizer` interface, `KimiVisionTextRecognizer` (only implementation), and `AiocrService`. Simplify all factory functions (`NewVisionTextRecognizerFromEnv`, `NewDebtDetailOCRRecognizerFromEnv`) to unconditionally return a Kimi recognizer. Update tests to reflect the single-provider world.

**Tech Stack:** Go, Kratos framework, `github.com/sashabaranov/go-openai` (OpenAI-compatible SDK, used by Kimi)

---

## File Structure

| Action | File | Role |
|--------|------|------|
| Modify | `blog/internal/service/aiocr.go` | Remove Ark, Fallback, Paddle references; keep only Kimi + interface + service |
| Modify | `blog/internal/service/aiocr_test.go` | Rewrite tests for single-provider world |
| Modify | `blog/internal/service/debt_detail_ocr_test.go` | Update recognizer env tests |
| Delete | `blog/internal/service/paddleocr.go` | PaddleOCR sidecar/subprocess recognizer (entire file) |
| Delete | `blog/internal/service/paddleocr_test.go` | PaddleOCR tests (entire file) |
| Delete | `blog/paddleocr_wrapper.bat` | Windows batch wrapper for PaddleOCR CLI |
| Delete | `blog/internal/service/__pycache__/` | Compiled Python cache (stale) |
| Regenerate | `blog/cmd/blog/wire_gen.go` | Wire DI code (auto-generated) |

Files intentionally NOT changed:
- `blog/internal/service/debt_detail_ocr.go` — uses `VisionTextRecognizer` interface, no provider-specific code
- `blog/internal/service/debt_detail_ocr_parser.go` — pure text parsing, no OCR dependency
- `blog/internal/service/debt_detail_ocr_parser_test.go` — parser tests, no OCR dependency
- `blog/internal/service/debtdetail.go` — uses `VisionTextRecognizer` interface via `NewDebtDetailOCRRecognizerFromEnv()`
- `blog/internal/service/service.go` — Wire `ProviderSet`, no changes needed (`NewAiocrService` still exists)
- `blog/api/ocr/v1/*` — protobuf definitions, unchanged

---

### Task 1: Delete PaddleOCR files and stale artifacts

**Files:**
- Delete: `blog/internal/service/paddleocr.go`
- Delete: `blog/internal/service/paddleocr_test.go`
- Delete: `blog/paddleocr_wrapper.bat`
- Delete: `blog/internal/service/__pycache__/` (entire directory)

- [ ] **Step 1: Delete the files**

```powershell
Remove-Item -Force "D:\code\blog\blog\internal\service\paddleocr.go"
Remove-Item -Force "D:\code\blog\blog\internal\service\paddleocr_test.go"
Remove-Item -Force "D:\code\blog\blog\paddleocr_wrapper.bat"
Remove-Item -Recurse -Force "D:\code\blog\blog\internal\service\__pycache__"
```

- [ ] **Step 2: Verify files are gone**

Run: `Get-ChildItem -Path "D:\code\blog\blog" -Recurse -Filter "*paddleocr*" -ErrorAction SilentlyContinue`
Expected: No output (all PaddleOCR files deleted)

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: remove PaddleOCR provider and stale artifacts

Delete paddleocr.go, paddleocr_test.go, paddleocr_wrapper.bat, and
__pycache__ directory. These are being removed as part of simplifying
the OCR module to use only the Kimi provider.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Simplify aiocr.go — remove Ark and Fallback, keep only Kimi

**Files:**
- Modify: `blog/internal/service/aiocr.go` (rewrite entire file)

- [ ] **Step 1: Replace aiocr.go with the simplified single-provider version**

Write the entire file (`D:\code\blog\blog\internal\service\aiocr.go`):

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	pb "blog/api/ocr/v1"

	ark "github.com/sashabaranov/go-openai"
)

const defaultKimiOCRModel = "kimi-k2.6"
const defaultKimiBaseURL = "https://api.moonshot.ai/v1"

// VisionTextRecognizer is the interface for OCR text recognition.
// Currently only implemented by KimiVisionTextRecognizer.
type VisionTextRecognizer interface {
	RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error)
}

// KimiVisionTextRecognizer uses the Kimi (Moonshot) API for vision-based text recognition.
type KimiVisionTextRecognizer struct {
	client *ark.Client
	model  string
}

// NewKimiVisionTextRecognizer creates a Kimi recognizer with the given API key.
// The model can be overridden via the KIMI_OCR_MODEL environment variable.
func NewKimiVisionTextRecognizer(apiKey string) *KimiVisionTextRecognizer {
	config := ark.DefaultConfig(strings.TrimSpace(apiKey))
	config.BaseURL = defaultKimiBaseURL
	model := strings.TrimSpace(os.Getenv("KIMI_OCR_MODEL"))
	if model == "" {
		model = defaultKimiOCRModel
	}
	return &KimiVisionTextRecognizer{
		client: ark.NewClientWithConfig(config),
		model:  model,
	}
}

// RecognizeText sends an image URL and prompt to the Kimi API and returns the recognized text.
func (r *KimiVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
	resp, err := r.client.CreateChatCompletion(
		ctx,
		ark.ChatCompletionRequest{
			Model: r.model,
			Messages: []ark.ChatCompletionMessage{
				{
					Role: ark.ChatMessageRoleUser,
					MultiContent: []ark.ChatMessagePart{
						{
							Type: ark.ChatMessagePartTypeImageURL,
							ImageURL: &ark.ChatMessageImageURL{
								URL: imageURL,
							},
						},
						{
							Type: ark.ChatMessagePartTypeText,
							Text: prompt,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("ocr returned no choices")
	}
	return resp.Choices[0].Message.Content, nil
}

// AiocrService implements the AiocrServer gRPC/HTTP service.
type AiocrService struct {
	recognizer VisionTextRecognizer
	pb.UnimplementedAiocrServer
}

// NewAiocrService creates an AiocrService using the recognizer from environment config.
func NewAiocrService() *AiocrService {
	return NewAiocrServiceWithRecognizer(NewVisionTextRecognizerFromEnv())
}

// NewAiocrServiceWithRecognizer creates an AiocrService with an explicit recognizer (for testing).
func NewAiocrServiceWithRecognizer(recognizer VisionTextRecognizer) *AiocrService {
	if recognizer == nil {
		recognizer = NewVisionTextRecognizerFromEnv()
	}
	return &AiocrService{recognizer: recognizer}
}

// Ocr handles OCR requests by delegating to the configured VisionTextRecognizer.
func (s *AiocrService) Ocr(ctx context.Context, req *pb.OcrRequest) (*pb.OcrReply, error) {
	// Use a fresh context with generous timeout, not the HTTP request context
	// (Kratos' http.Timeout may cancel the request context prematurely)
	ocrCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := s.recognizer.RecognizeText(ocrCtx, req.ImgBaseStr, "直接输出图片内容，不要输出其他东西")
	if err != nil {
		return nil, fmt.Errorf("ocr failed: %w", err)
	}
	return &pb.OcrReply{Res: res}, nil
}

// NewVisionTextRecognizerFromEnv reads KIMI_API_KEY (with MOONSHOT_API_KEY fallback)
// and returns a KimiVisionTextRecognizer.
func NewVisionTextRecognizerFromEnv() VisionTextRecognizer {
	apiKey := strings.TrimSpace(os.Getenv("KIMI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("MOONSHOT_API_KEY"))
	}
	return NewKimiVisionTextRecognizer(apiKey)
}

// NewDebtDetailOCRRecognizerFromEnv returns a Kimi recognizer for debt detail OCR.
// Previously supported a fallback chain; now always returns Kimi.
func NewDebtDetailOCRRecognizerFromEnv() VisionTextRecognizer {
	return NewVisionTextRecognizerFromEnv()
}
```

- [ ] **Step 2: Verify the file has no references to removed types**

Run: `Select-String -Path "D:\code\blog\blog\internal\service\aiocr.go" -Pattern "Ark|Fallback|Paddle|paddleocr|defaultOCRModel|defaultArkBaseURL|defaultOCRProvider|defaultDebtDetailOCRProvider|defaultDebtDetailFallbackOCRProvider|defaultPaddleOCRCommand"`

Expected: No matches

- [ ] **Step 3: Commit**

```bash
git add blog/internal/service/aiocr.go
git commit -m "refactor: remove Ark and Fallback OCR providers, keep only Kimi

Removed ArkVisionTextRecognizer, FallbackVisionTextRecognizer, and all
PaddleOCR references from aiocr.go. Simplified factory functions to
always return KimiVisionTextRecognizer.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Rewrite aiocr_test.go for single-provider world

**Files:**
- Modify: `blog/internal/service/aiocr_test.go` (rewrite entire file)

- [ ] **Step 1: Replace aiocr_test.go with tests that only cover Kimi**

Write the entire file (`D:\code\blog\blog\internal\service\aiocr_test.go`):

```go
package service

import (
	"context"
	"testing"

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

func TestNewDebtDetailOCRRecognizerFromEnv_ReturnsKimi(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "kimi-test-key")

	recognizer := NewDebtDetailOCRRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "NewDebtDetailOCRRecognizerFromEnv should return Kimi")
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
```

Note: The test file needs to import `pb "blog/api/ocr/v1"` for the last two tests to compile. Add it to the import block.

- [ ] **Step 2: Verify the test file has no references to removed types**

Run: `Select-String -Path "D:\code\blog\blog\internal\service\aiocr_test.go" -Pattern "PaddleOCR|ArkVision|Fallback|SelectsArk|SelectsPaddle|SelectsMoonshotAlias|DefaultsToPaddle"`

Expected: No matches

- [ ] **Step 3: Commit**

```bash
git add blog/internal/service/aiocr_test.go
git commit -m "test: rewrite aiocr tests for single-provider (Kimi-only) world

Removed tests for Ark, Paddle, and Fallback recognizers. Added tests
for KimiVisionTextRecognizer creation, env var handling, and AiocrService
delegation.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: Update debt_detail_ocr_test.go — fix recognizer env tests

**Files:**
- Modify: `blog/internal/service/debt_detail_ocr_test.go:61-84`

- [ ] **Step 1: Replace the two OCR-provider-specific test functions**

The functions to replace are `TestNewDebtDetailOCRRecognizerFromEnv_DefaultsToKimiThenPaddleOCR` (lines 61-75) and `TestNewDebtDetailOCRRecognizerFromEnv_ExplicitArk` (lines 77-84).

Remove both old functions. Add one replacement function after the `newDebtDetailOCRTestService` helper (after line 59):

```go
func TestNewDebtDetailOCRRecognizerFromEnv_ReturnsKimi(t *testing.T) {
	t.Setenv("KIMI_API_KEY", "kimi-test-key")

	recognizer := NewDebtDetailOCRRecognizerFromEnv()

	_, ok := recognizer.(*KimiVisionTextRecognizer)
	assert.True(t, ok, "NewDebtDetailOCRRecognizerFromEnv should return KimiVisionTextRecognizer")
}
```

- [ ] **Step 2: Verify the old tests are fully removed**

Run: `Select-String -Path "D:\code\blog\blog\internal\service\debt_detail_ocr_test.go" -Pattern "DefaultsToKimiThenPaddleOCR|ExplicitArk|PaddleOCRTextRecognizer|ArkVisionTextRecognizer|FallbackVisionTextRecognizer"`

Expected: No matches

- [ ] **Step 3: Commit**

```bash
git add blog/internal/service/debt_detail_ocr_test.go
git commit -m "test: update debt detail OCR tests for Kimi-only provider

Replaced TestNewDebtDetailOCRRecognizerFromEnv_DefaultsToKimiThenPaddleOCR
and TestNewDebtDetailOCRRecognizerFromEnv_ExplicitArk with a single test
that verifies Kimi is always returned.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: Verify compilation

**Files:**
- No changes — verification only

- [ ] **Step 1: Run go build to verify compilation**

Run: `cd D:\code\blog\blog; go build ./...`
Expected: Exit code 0, no errors

- [ ] **Step 2: Run vet to catch any issues**

Run: `cd D:\code\blog\blog; go vet ./...`
Expected: Exit code 0, no warnings

---

### Task 6: Run all unit tests

**Files:**
- No changes — verification only

- [ ] **Step 1: Run the service package tests**

Run: `cd D:\code\blog\blog; go test ./internal/service/... -v -count=1`
Expected: All tests PASS, no failures

- [ ] **Step 2: Run all project tests**

Run: `cd D:\code\blog\blog; go test ./... -count=1`
Expected: All tests PASS, no failures

- [ ] **Step 3: Commit if any generated files changed**

Check: `git status`
If `wire_gen.go` or any other files changed, commit them:

```bash
git add blog/cmd/blog/wire_gen.go
git commit -m "chore: regenerate Wire after OCR provider cleanup

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 7: Regenerate Wire dependency injection

**Files:**
- Regenerate: `blog/cmd/blog/wire_gen.go`

- [ ] **Step 1: Run Wire code generation**

Run: `cd D:\code\blog\blog\cmd\blog; wire`
Expected: Exit code 0, output shows "wire: blog/cmd/blog: wrote ..."

- [ ] **Step 2: Verify the generated file still references NewAiocrService**

Run: `Select-String -Path "D:\code\blog\blog\cmd\blog\wire_gen.go" -Pattern "NewAiocrService"`
Expected: One match showing `aiocrService := service.NewAiocrService()`

- [ ] **Step 3: Commit**

```bash
git add blog/cmd/blog/wire_gen.go
git commit -m "chore: regenerate Wire after OCR provider simplification

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 8: Full integration verification

**Files:**
- No changes — verification only

- [ ] **Step 1: Build the binary**

Run: `cd D:\code\blog\blog; go build -o bin/blog.exe ./cmd/blog`
Expected: Exit code 0, binary created at `bin/blog.exe`

- [ ] **Step 2: Start the backend service**

Run: `cd D:\code\blog\blog; make run`
Expected: Service starts without errors. Watch for any panic or OCR-related error messages.

- [ ] **Step 3: Test the OCR endpoint with curl**

Run:
```powershell
$body = '{"imgBaseStr": "data:image/png;base64,iVBORw0KGgo="}'
Invoke-RestMethod -Uri "http://localhost:8000/ocr/aiocr" -Method Post -Body $body -ContentType "application/json"
```

Expected: Returns a JSON response (may fail with "ocr failed" if no valid Kimi API key, but should NOT panic or complain about missing providers)

- [ ] **Step 4: Stop the service** (Ctrl+C in the make run terminal)

- [ ] **Step 5: Final check — git status clean**

Run: `git status`
Expected: Only expected changes, no untracked files from the work.

---

## Self-Review Checklist

1. **Spec coverage:** The requirement is "remove all OCR methods except Kimi." Every removed provider is covered:
   - PaddleOCR → Task 1 deletes paddleocr.go, paddleocr_test.go, paddleocr_wrapper.bat
   - Ark/Doubao → Task 2 removes ArkVisionTextRecognizer from aiocr.go
   - Fallback → Task 2 removes FallbackVisionTextRecognizer from aiocr.go
   - Tests updated in Tasks 3 and 4
   - Wire regeneration in Task 7
   - Integration verification in Task 8

2. **Placeholder scan:** No TBD, TODO, "add error handling," or other placeholder patterns found.

3. **Type consistency:**
   - `VisionTextRecognizer` interface is preserved (used by debt_detail_ocr.go and debtdetail.go)
   - `KimiVisionTextRecognizer` is preserved as the sole implementation
   - `NewAiocrService()` and `NewAiocrServiceWithRecognizer()` signatures unchanged (used by Wire)
   - `NewDebtDetailOCRRecognizerFromEnv()` return type unchanged (`VisionTextRecognizer`)
   - `NewVisionTextRecognizerFromEnv()` return type unchanged (`VisionTextRecognizer`)
   - No function renames that would break external callers
