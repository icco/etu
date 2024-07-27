package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jomei/notionapi"
)

type Post struct {
	ID         string
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
	cache, err := c.cacheFromFile()
	if err == nil {
		if time.Since(cache.Saved) < 10*time.Minute {
			return cache.Duration, nil
		}
	}

	posts, err := c.ListPosts(ctx, 1)
	if err != nil {
		return time.Duration(0), err
	}

	if len(posts) == 0 {
		return time.Duration(0), fmt.Errorf("no posts found")
	}
	dur := time.Since(posts[0].CreatedAt)
	if err := c.cacheToFile(dur); err != nil {
		return time.Duration(0), err
	}

	return dur, nil
}

type cacheData struct {
	Saved    time.Time
	Duration time.Duration
}

func (c *Config) cacheToFile(dur time.Duration) error {
	f, err := os.Create("/tmp/etu.cache")
	if err != nil {
		return err
	}
	defer f.Close()

	data := cacheData{
		Saved:    time.Now(),
		Duration: dur,
	}

	// Create an encoder and send a value.
	enc := gob.NewEncoder(f)
	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}

func (c *Config) cacheFromFile() (cacheData, error) {
	var data cacheData
	f, err := os.Open("/tmp/etu.cache")
	if err != nil {
		return data, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return data, err
	}

	return data, nil
}

func (c *Config) SaveEntry(ctx context.Context, text string) error {
	post := &Post{
		Text: text,
		ID:   uuid.New().String(),
	}

	dbID, err := c.getDatabaseID(ctx)
	if err != nil {
		return err
	}

	client := c.GetClient()
	if _, err := client.Page.Create(ctx, &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: dbID,
		},
		Properties: notionapi.Properties{
			"ID": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{Text: &notionapi.Text{Content: post.ID}},
				},
			},
		},
		Children: ToBlocks(post.Text),
	}); err != nil {
		return err
	}

	return nil
}

func ToBlocks(text string) []notionapi.Block {
	var blocks []notionapi.Block
	for _, line := range strings.Split(text, "\n") {
		block := &notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					{Text: &notionapi.Text{Content: line}},
				},
			},
		}
		block.Type = notionapi.BlockTypeParagraph
		block.Object = notionapi.ObjectTypeBlock

		blocks = append(blocks, block)
	}

	return blocks
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
		PageSize: count,
	})

	if err != nil {
		return nil, err
	}

	var ret []*Post
	for _, page := range resp.Results {
		rawTags := page.Properties["Tags"]
		tagData, ok := rawTags.(*notionapi.MultiSelectProperty)
		if !ok {
			return nil, fmt.Errorf("tags property is not a multi-select: %+v", rawTags)
		}
		var tags []string
		for _, tag := range tagData.MultiSelect {
			tags = append(tags, tag.Name)
		}

		rawID := page.Properties["ID"]
		idData, ok := rawID.(*notionapi.TitleProperty)
		if !ok {
			return nil, fmt.Errorf("id property is not a title: %+v", rawID)
		}
		id := idData.Title[0].PlainText

		blockResp, err := client.Block.GetChildren(ctx, notionapi.BlockID(page.ID), &notionapi.Pagination{PageSize: 10})
		if err != nil {
			return nil, err
		}

		text := ""
		for _, block := range blockResp.Results {
			switch block.GetType() {
			case notionapi.BlockTypeParagraph:
				paragraph, ok := block.(*notionapi.ParagraphBlock)
				if !ok {
					return nil, fmt.Errorf("paragraph is incorrect block type: %+v", block)
				}
				text += paragraph.GetRichTextString() + "\n"
			default:
				fmt.Printf("skipping block type: %s\n", block.GetType())
			}
		}

		text = strings.TrimSpace(text)

		ret = append(ret, &Post{
			ID:         id,
			Tags:       tags,
			Text:       text,
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
