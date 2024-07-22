package client

import (
	"context"
	"fmt"
	"time"
)

type Post struct {
	Title      string
	Tags       []string
	Text       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	return time.Duration(0), fmt.Errorf("not implemented")
}

func SaveEntry(ctx context.Context, text string) error {
	return fmt.Errorf("not implemented")
}

func DeletePost(ctx context.Context, key string) error {
	return fmt.Errorf("not implemented")
}

func GetPost(ctx context.Context, key string) (*Post, error) {
	return nil, fmt.Errorf("not implemented")
}

func ListPosts(ctx context.Context, count int) ([]*Post, error) {
	return nil, fmt.Errorf("not implemented")
}
