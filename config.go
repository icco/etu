package main

import (
	"fmt"
	"os"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	NotionKey    string
	OpenAIAPIKey string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		NotionKey:    os.Getenv("NOTION_KEY"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

// Validate checks that all required configuration values are present.
func (c *Config) Validate() error {
	if c.NotionKey == "" {
		return fmt.Errorf("NOTION_KEY is required")
	}

	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}

	return nil
}
