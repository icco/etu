package ai

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func GenerateTags(ctx context.Context, text string) ([]string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("given the journal entry %q, generate three words to summarize the content.", text),
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini20240718,
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
