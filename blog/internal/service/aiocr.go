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
