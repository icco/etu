package ai

import (
	"context"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func GenerateTags(ctx context.Context, text string) ([]string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4,
		Messages: messages,
	}

	var tags []string
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	log.Println(resp)

	return tags, nil
}
