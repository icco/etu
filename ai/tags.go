package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

var (
	openAIClient *openai.Client
	clientOnce   sync.Once
)

func getOpenAIClient() *openai.Client {
	clientOnce.Do(func() {
		openAIClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	})
	return openAIClient
}

// GenerateTags generates a list of tags for a given text using OpenAI.
func GenerateTags(ctx context.Context, text string) ([]string, error) {
	client := getOpenAIClient()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("given the journal entry %q, generate a few options of single words to summarize the content. Output should be a comma separated list. You should only output the list of tags, no other text. You should only output three tags maximum.", text),
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT5Nano,
		Messages: messages,
	}

	var tags []string
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, choice := range resp.Choices {
		outText := choice.Message.Content
		newTags := strings.Split(outText, ",")
		for _, tag := range newTags {
			tags = append(tags, strings.TrimSpace(tag))
		}
	}
	// log.Printf("tags: %+v", tags)

	return tags, nil
}
