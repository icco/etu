package client

import (
	"context"
	"fmt"
	"time"

	"github.com/jomei/notionapi"
)

type Post struct {
	Title      string
	Tags       []string
	Text       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type Config struct {
	key      string
	rootPage string
}

func New(key string) (*Config, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	return &Config{
		key:      key,
		rootPage: "Journal",
	}, nil
}

func (c *Config) GetClient() *notionapi.Client {
	return notionapi.NewClient(c.key)
}

func (c *Config) TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	return time.Duration(0), fmt.Errorf("not implemented")
}

func (c *Config) SaveEntry(ctx context.Context, text string) error {
	return fmt.Errorf("not implemented")
}

func (c *Config) DeletePost(ctx context.Context, key string) error {
	return fmt.Errorf("not implemented")
}

func (c *Config) GetPost(ctx context.Context, key string) (*Post, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Config) ListPosts(ctx context.Context, count int) ([]*Post, error) {

	return nil, fmt.Errorf("not implemented")
}
