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

const defaultOCRModel = "doubao-seed-1-6-251015"
const defaultArkBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
const defaultKimiOCRModel = "kimi-k2.6"
const defaultKimiBaseURL = "https://api.moonshot.ai/v1"
const defaultOCRProvider = "kimi"
const defaultDebtDetailOCRProvider = "kimi"
const defaultDebtDetailFallbackOCRProvider = "paddle"
const defaultPaddleOCRCommand = "paddleocr"

type VisionTextRecognizer interface {
	RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error)
}

type FallbackVisionTextRecognizer struct {
	primary   VisionTextRecognizer
	secondary VisionTextRecognizer
}

func NewFallbackVisionTextRecognizer(primary VisionTextRecognizer, secondary VisionTextRecognizer) *FallbackVisionTextRecognizer {
	return &FallbackVisionTextRecognizer{
		primary:   primary,
		secondary: secondary,
	}
}

func (r *FallbackVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
	if r == nil {
		return "", errors.New("ocr recognizer unavailable")
	}
	if r.primary != nil {
		if result, err := r.primary.RecognizeText(ctx, imageURL, prompt); err == nil {
			return result, nil
		}
	}
	if r.secondary != nil {
		return r.secondary.RecognizeText(ctx, imageURL, prompt)
	}
	return "", errors.New("ocr recognizer unavailable")
}

type ArkVisionTextRecognizer struct {
	client *ark.Client
	model  string
}

func NewArkVisionTextRecognizer(apiKey string) *ArkVisionTextRecognizer {
	config := ark.DefaultConfig(apiKey)
	config.BaseURL = defaultArkBaseURL
	return &ArkVisionTextRecognizer{
		client: ark.NewClientWithConfig(config),
		model:  defaultOCRModel,
	}
}

func (r *ArkVisionTextRecognizer) RecognizeText(ctx context.Context, imageURL string, prompt string) (string, error) {
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
			ReasoningEffort: "medium",
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

type KimiVisionTextRecognizer struct {
	client *ark.Client
	model  string
}

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

type AiocrService struct {
	recognizer VisionTextRecognizer
	pb.UnimplementedAiocrServer
}

func NewAiocrService() *AiocrService {
	return NewAiocrServiceWithRecognizer(NewVisionTextRecognizerFromEnv())
}

func NewAiocrServiceWithRecognizer(recognizer VisionTextRecognizer) *AiocrService {
	if recognizer == nil {
		recognizer = NewVisionTextRecognizerFromEnv()
	}
	return &AiocrService{recognizer: recognizer}
}

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

func NewVisionTextRecognizerFromEnv() VisionTextRecognizer {
	provider := strings.TrimSpace(os.Getenv("OCR_PROVIDER"))
	if provider == "" {
		provider = "paddle"
	}
	return newVisionTextRecognizer(provider)
}

func NewDebtDetailOCRRecognizerFromEnv() VisionTextRecognizer {
	provider := strings.TrimSpace(os.Getenv("OCR_PROVIDER"))
	if provider == "" {
		return NewFallbackVisionTextRecognizer(
			newVisionTextRecognizer(defaultDebtDetailOCRProvider),
			newVisionTextRecognizer(defaultDebtDetailFallbackOCRProvider),
		)
	}
	return newVisionTextRecognizer(provider)
}

func newVisionTextRecognizer(provider string) VisionTextRecognizer {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "paddle", "paddleocr":
		command := strings.TrimSpace(os.Getenv("PADDLE_OCR_COMMAND"))
		if command == "" {
			command = defaultPaddleOCRCommand
		}
		return NewPaddleOCRTextRecognizer(command)
	case "kimi", "moonshot":
		apiKey := strings.TrimSpace(os.Getenv("KIMI_API_KEY"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(os.Getenv("MOONSHOT_API_KEY"))
		}
		return NewKimiVisionTextRecognizer(apiKey)
	case "ark", "":
		return NewArkVisionTextRecognizer("")
	default:
		return NewArkVisionTextRecognizer("")
	}
}
