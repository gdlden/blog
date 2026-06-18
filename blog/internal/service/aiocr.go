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

const defaultDeepSeekOCRModel = "deepseek-vl2"
const defaultDeepSeekBaseURL = "https://api.deepseek.com"

// VisionTextRecognizer is the interface for OCR text recognition.
// Currently implemented by DeepSeekVisionTextRecognizer.
type VisionTextRecognizer interface {
	RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error)
}

// DeepSeekVisionTextRecognizer uses the DeepSeek API for vision-based text recognition.
type DeepSeekVisionTextRecognizer struct {
	client *ark.Client
	model  string
}

// NewDeepSeekVisionTextRecognizer creates a DeepSeek recognizer with the given API key.
// The model can be overridden via the OCR_MODEL environment variable.
func NewDeepSeekVisionTextRecognizer(apiKey string) *DeepSeekVisionTextRecognizer {
	config := ark.DefaultConfig(strings.TrimSpace(apiKey))
	config.BaseURL = defaultDeepSeekBaseURL
	model := strings.TrimSpace(os.Getenv("OCR_MODEL"))
	if model == "" {
		model = defaultDeepSeekOCRModel
	}
	return &DeepSeekVisionTextRecognizer{
		client: ark.NewClientWithConfig(config),
		model:  model,
	}
}

// RecognizeText sends an image URL and prompt to the DeepSeek API and returns the recognized text.
func (r *DeepSeekVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
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

// NewVisionTextRecognizerFromEnv reads OCR_API_KEY (with KIMI_API_KEY/MOONSHOT_API_KEY fallback)
// and returns a DeepSeekVisionTextRecognizer.
func NewVisionTextRecognizerFromEnv() VisionTextRecognizer {
	apiKey := strings.TrimSpace(os.Getenv("OCR_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("KIMI_API_KEY"))
	}
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("MOONSHOT_API_KEY"))
	}
	return NewDeepSeekVisionTextRecognizer(apiKey)
}

// NewDebtDetailOCRRecognizerFromEnv returns a DeepSeek recognizer for debt detail OCR.
// Delegates to NewVisionTextRecognizerFromEnv.
func NewDebtDetailOCRRecognizerFromEnv() VisionTextRecognizer {
	return NewVisionTextRecognizerFromEnv()
}
