package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/icco/etu/ai"
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
	// TODO: figure out timeouts
	return notionapi.NewClient(
		notionapi.Token(c.key),
		notionapi.WithVersion("2022-06-28"),
		notionapi.WithRetry(2),
	)
}

func (c *Config) UpdateCache(ctx context.Context) error {
	posts, err := c.ListPosts(ctx, 1)
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		return fmt.Errorf("no posts found")
	}
	dur := time.Since(posts[0].CreatedAt)
	if err := c.cacheToFile(dur); err != nil {
		return err
	}

	return nil
}

func (c *Config) TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	if cache, _ := c.cacheFromFile(); cache != nil {
		return cache.Duration, nil
	}

	if err := c.UpdateCache(ctx); err != nil {
		return time.Duration(0), fmt.Errorf("updating cache %w", err)
	}

	if cache, _ := c.cacheFromFile(); cache != nil {
		return cache.Duration, nil
	}

	return time.Duration(0), fmt.Errorf("cache still not found")
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

func (c *Config) cacheFromFile() (*cacheData, error) {
	data := &cacheData{}
	f, err := os.Open("/tmp/etu.cache")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	if err := dec.Decode(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Config) SaveEntry(ctx context.Context, text string) error {
	post := &Post{
		Text: text,
		ID:   uuid.New().String(),
	}

	tags, err := ai.GenerateTags(ctx, text)
	if err != nil {
		return err
	}

	dbID, err := c.getDatabaseID(ctx)
	if err != nil {
		return err
	}

	tagOptions := make([]notionapi.Option, len(tags))
	for i, tag := range tags {
		tagOptions[i] = notionapi.Option{Name: tag}
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
			"Tags": notionapi.MultiSelectProperty{
				MultiSelect: tagOptions,
			},
		},
		Children: ToBlocks(post.Text),
	}); err != nil {
		return err
	}

	if err := c.UpdateCache(ctx); err != nil {
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

	return c.processPages(ctx, client, resp.Results)
}

// SearchPosts uses Notion's native Search API to find matching pages,
// then fetches content only for those matches. This is much faster than
// fetching all pages and searching them locally.
func (c *Config) SearchPosts(ctx context.Context, query string, maxResults int) ([]*Post, error) {
	dbID, err := c.getDatabaseID(ctx)
	if err != nil {
		return nil, err
	}

	client := c.GetClient()
	
	// If query is empty, just return recent posts
	if query == "" {
		return c.ListPosts(ctx, maxResults)
	}

	// Use Notion's Search API to find matching pages
	// Search within the database by filtering for pages in this database
	searchResp, err := client.Search.Do(ctx, &notionapi.SearchRequest{
		Query: query,
		Filter: notionapi.SearchFilter{
			Value:    "page",
			Property: "object",
		},
		PageSize: maxResults,
	})
	if err != nil {
		return nil, err
	}

	// Filter results to only include pages from our database
	var matchingPages []notionapi.Page
	for _, result := range searchResp.Results {
		if page, ok := result.(*notionapi.Page); ok {
			// Check if this page belongs to our database
			if page.Parent.Type == notionapi.ParentTypeDatabaseID && page.Parent.DatabaseID == notionapi.DatabaseID(dbID) {
				matchingPages = append(matchingPages, *page)
			}
		}
	}

	// If no results from search API, fall back to database query with text filter
	if len(matchingPages) == 0 {
		return c.searchViaDatabaseQuery(ctx, dbID, query, maxResults)
	}

	// Fetch content only for matching pages
	return c.processPages(ctx, client, matchingPages)
}

// searchViaDatabaseQuery is a fallback that queries the database directly
// when Search API doesn't return results (e.g., for very specific queries)
func (c *Config) searchViaDatabaseQuery(ctx context.Context, dbID notionapi.DatabaseID, query string, maxResults int) ([]*Post, error) {
	client := c.GetClient()
	var results []*Post
	var cursor notionapi.Cursor
	
	// Search incrementally - fetch pages and search them
	for len(results) < maxResults {
		req := &notionapi.DatabaseQueryRequest{
			Sorts: []notionapi.SortObject{
				{Property: "Created At", Direction: notionapi.SortOrderDESC},
			},
			PageSize: 100, // Max page size
		}
		if cursor != "" {
			req.StartCursor = cursor
		}

		resp, err := client.Database.Query(ctx, dbID, req)
		if err != nil {
			return nil, err
		}

		posts, err := c.processPages(ctx, client, resp.Results)
		if err != nil {
			return nil, err
		}

		// Search this batch
		for _, post := range posts {
			if c.matchesQuery(post, query) {
				results = append(results, post)
				if len(results) >= maxResults {
					return results, nil
				}
			}
		}

		// Check if there are more pages
		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

// matchesQuery performs a simple string match check
func (c *Config) matchesQuery(post *Post, query string) bool {
	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(post.Text)
	tagsLower := strings.ToLower(strings.Join(post.Tags, " "))
	
	// Check if query appears in text or tags
	return strings.Contains(textLower, queryLower) || strings.Contains(tagsLower, queryLower)
}

func (c *Config) processPages(ctx context.Context, client *notionapi.Client, pages []notionapi.Page) ([]*Post, error) {
	ret := make([]*Post, 0, len(pages))
	
	for _, page := range pages {
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

		// Fetch blocks with pagination to get all content
		text := ""
		var cursor string
		for {
			pagination := &notionapi.Pagination{PageSize: 100}
			if cursor != "" {
				pagination.StartCursor = notionapi.Cursor(cursor)
			}
			
			blockResp, err := client.Block.GetChildren(ctx, notionapi.BlockID(page.ID), pagination)
			if err != nil {
				return nil, err
			}

			for _, block := range blockResp.Results {
				switch block.GetType() {
				case notionapi.BlockTypeParagraph:
					paragraph, ok := block.(*notionapi.ParagraphBlock)
					if !ok {
						return nil, fmt.Errorf("paragraph is incorrect block type: %+v", block)
					}
					text += paragraph.GetRichTextString() + "\n"
				default:
					// Silently skip other block types
				}
			}

			if !blockResp.HasMore {
				break
			}
			cursor = blockResp.NextCursor
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
