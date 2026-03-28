package service

import (
	"context"
	"fmt"

	pb "blog/api/ocr/v1"

	ark "github.com/sashabaranov/go-openai"
)

type AiocrService struct {
	pb.UnimplementedAiocrServer
}

func NewAiocrService() *AiocrService {
	return &AiocrService{}
}

func (s *AiocrService) Ocr(ctx context.Context, req *pb.OcrRequest) (*pb.OcrReply, error) {
	// (os.Getenv(""))
	config := ark.DefaultConfig("")
	config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	client := ark.NewClientWithConfig(config)

	fmt.Println("----- image input request -----")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		ark.ChatCompletionRequest{
			Model: "doubao-seed-1-6-251015",
			Messages: []ark.ChatCompletionMessage{
				{
					Role: ark.ChatMessageRoleUser,
					MultiContent: []ark.ChatMessagePart{
						{
							Type: ark.ChatMessagePartTypeImageURL,
							ImageURL: &ark.ChatMessageImageURL{
								URL: req.ImgBaseStr,
							},
						},
						{
							Type: ark.ChatMessagePartTypeText,
							Text: "直接输出图片内容，不要输出其他东西",
						},
					},
				},
			},
			ReasoningEffort: "medium",
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		// return
	}
	if resp.Choices[0].Message.ReasoningContent != "" {
		// fmt.Println(resp.Choices[0].Message.ReasoningContent)
	}
	res := resp.Choices[0].Message.Content

	return &pb.OcrReply{
		Res: res,
	}, nil
}
