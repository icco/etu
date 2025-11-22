package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

var (
	openAIClient *openai.Client
	clientOnce   sync.Once
	apiKey       string
)

func getOpenAIClient(key string) *openai.Client {
	clientOnce.Do(func() {
		apiKey = key
		if apiKey != "" {
			openAIClient = openai.NewClient(apiKey)
		}
	})
	return openAIClient
}

// GenerateTags generates a list of tags for a given text using OpenAI.
// If apiKey is empty, returns empty tags without error.
func GenerateTags(ctx context.Context, text string, apiKey string) ([]string, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("no OpenAI key configured")
	}

	client := getOpenAIClient(apiKey)

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
