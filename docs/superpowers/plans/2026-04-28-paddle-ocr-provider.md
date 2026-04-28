# Paddle OCR Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a low-cost PaddleOCR backend provider for debt detail OCR while keeping the existing Ark visual model available.

**Architecture:** Keep the current `VisionTextRecognizer` interface. Add a command-backed PaddleOCR recognizer that writes uploaded images to a temp file, invokes a configurable OCR command, returns raw text, and lets the existing debt detail parser handle structured extraction.

**Tech Stack:** Go, Kratos service layer, `os/exec`, existing Go tests.

---

### Task 1: PaddleOCR Provider Selection

**Files:**
- Modify: `blog/internal/service/aiocr.go`
- Test: `blog/internal/service/aiocr_test.go`

- [ ] Write tests for `NewVisionTextRecognizerFromEnv` selecting `ark`, defaulting to `ark`, and selecting `paddle`.
- [ ] Verify tests fail because the provider factory does not exist.
- [ ] Implement `NewVisionTextRecognizerFromEnv`, `newVisionTextRecognizer`, and `NewPaddleOCRTextRecognizer`.
- [ ] Verify tests pass.

### Task 2: Command-Backed PaddleOCR Recognizer

**Files:**
- Create: `blog/internal/service/paddleocr.go`
- Test: `blog/internal/service/paddleocr_test.go`

- [ ] Write tests for extracting base64 image data to a temp file and invoking a fake command.
- [ ] Write tests for command failures returning a useful error.
- [ ] Verify tests fail because the recognizer does not exist.
- [ ] Implement command invocation with `context.Context`, temp-file cleanup, and text normalization.
- [ ] Verify tests pass.

### Task 3: Wire Debt Detail OCR to Provider Factory

**Files:**
- Modify: `blog/internal/service/aiocr.go`
- Modify: `blog/internal/service/debtdetail.go`
- Test: existing service tests
- Docs: `blog/FUNCTION_INDEX.md`

- [ ] Replace direct `NewArkVisionTextRecognizer("")` defaults with `NewVisionTextRecognizerFromEnv()`.
- [ ] Keep test constructors accepting injected recognizers unchanged.
- [ ] Update `blog/FUNCTION_INDEX.md` for added backend functions.
- [ ] Run `go test ./...`.
