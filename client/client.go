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
	return notionapi.NewClient(notionapi.Token(c.key), notionapi.WithVersion("2022-06-28"))
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
	dbID, err := c.getDatabaseID(ctx)
	if err != nil {
		return nil, err
	}

	client := c.GetClient()
	resp, err := client.Database.Query(ctx, dbID, &notionapi.DatabaseQueryRequest{
		Sorts: []notionapi.SortObject{
			{Property: "Created At", Direction: notionapi.SortOrderDESC},
		},
	})

	if err != nil {
		return nil, err
	}

	var ret []*Post
	for _, page := range resp.Results {
		tags := page.Properties["Tags"]
		id := page.Properties["ID"]
		fmt.Printf("tags: %+v\n", tags)
		fmt.Printf("id: %+v\n", id)

		blockResp, err := client.Block.GetChildren(ctx, notionapi.BlockID(page.ID), nil)
		if err != nil {
			return nil, err
		}
		fmt.Printf("blockResp: %+v\n", blockResp)

		ret = append(ret, &Post{
			Text:       page.GetObject().String(),
			CreatedAt:  page.CreatedTime,
			ModifiedAt: page.LastEditedTime,
		})
	}

	return ret, nil
}

func (c *Config) getDatabaseID(ctx context.Context) (notionapi.DatabaseID, error) {
	client := c.GetClient()
	resp, err := client.Search.Do(ctx, &notionapi.SearchRequest{
		Query: c.rootPage,
		Filter: notionapi.SearchFilter{
			Value:    "database",
			Property: "object",
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Results) == 0 {
		return "", fmt.Errorf("root page not found")
	}

	if len(resp.Results) > 1 {
		return "", fmt.Errorf("multiple root pages found")
	}

	db, ok := resp.Results[0].(*notionapi.Database)
	if !ok {
		return "", fmt.Errorf("root page is not a database")
	}

	id := notionapi.DatabaseID(db.ID.String())

	return id, nil
}
