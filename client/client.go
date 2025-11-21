package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/icco/etu/ai"
	"github.com/jomei/notionapi"
)

// Post represents a journal entry.
type Post struct {
	// ID is the unique identifier of the post.
	ID     string
	PageID string // Notion page ID for fetching full content
	// Tags are the tags associated with the post.
	Tags []string
	// Text is the content of the post.
	Text string
	// CreatedAt is the time the post was created.
	CreatedAt time.Time
	// ModifiedAt is the time the post was last modified.
	ModifiedAt time.Time
}

// Config holds the configuration for the client.
type Config struct {
	key        string
	rootPage   string
	cachedDbID notionapi.DatabaseID // Cache database ID to avoid repeated API calls
	client     *notionapi.Client    // Cached Notion client
	clientOnce sync.Once            // Ensures client is initialized only once
}

// New creates a new client configuration.
func New(key string) (*Config, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	return &Config{
		key:      key,
		rootPage: "Journal",
	}, nil
}

// GetClient returns a cached Notion client.
func (c *Config) GetClient() *notionapi.Client {
	c.clientOnce.Do(func() {
		// TODO: figure out timeouts
		c.client = notionapi.NewClient(
			notionapi.Token(c.key),
			notionapi.WithVersion("2022-06-28"),
			notionapi.WithRetry(2),
		)
	})
	return c.client
}

// UpdateCache updates the cache with the latest post.
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

// TimeSinceLastPost returns the time since the last post was created.
func (c *Config) TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	cache, _ := c.cacheFromFile()
	if cache != nil {
		// Use cache if it's less than 5 minutes old (avoids unnecessary API calls)
		if time.Since(cache.Saved) < 5*time.Minute {
			return cache.Duration, nil
		}
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

// SaveEntry saves a new journal entry to Notion.
func (c *Config) SaveEntry(ctx context.Context, text string) error {
	post := &Post{
		Text: text,
		ID:   uuid.New().String(),
	}

	// Run tag generation and database ID lookup in parallel
	type tagResult struct {
		tags []string
		err  error
	}
	type dbResult struct {
		dbID notionapi.DatabaseID
		err  error
	}

	tagChan := make(chan tagResult, 1)
	dbChan := make(chan dbResult, 1)

	// Generate tags in parallel
	go func() {
		tags, err := ai.GenerateTags(ctx, text)
		tagChan <- tagResult{tags: tags, err: err}
	}()

	// Get database ID in parallel
	go func() {
		dbID, err := c.getDatabaseID(ctx)
		dbChan <- dbResult{dbID: dbID, err: err}
	}()

	// Wait for both results
	tagRes := <-tagChan
	if tagRes.err != nil {
		return tagRes.err
	}

	dbRes := <-dbChan
	if dbRes.err != nil {
		return dbRes.err
	}

	tagOptions := make([]notionapi.Option, len(tagRes.tags))
	for i, tag := range tagRes.tags {
		tagOptions[i] = notionapi.Option{Name: tag}
	}

	client := c.GetClient()
	createdPage, err := client.Page.Create(ctx, &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: dbRes.dbID,
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
	})
	if err != nil {
		return err
	}

	// Update cache using the created page's timestamp (much faster than querying)
	dur := time.Since(createdPage.CreatedTime)
	if err := c.cacheToFile(dur); err != nil {
		// Don't fail the save if cache update fails
		_ = err
	}

	return nil
}

// ToBlocks converts a string of text to a list of Notion blocks.
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

// DeletePost deletes a journal entry.
func (c *Config) DeletePost(ctx context.Context, key string) error {
	return fmt.Errorf("not implemented")
}

// GetPost retrieves a journal entry by its key.
func (c *Config) GetPost(ctx context.Context, key string) (*Post, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListPosts lists the most recent journal entries.
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

// processPages processes pages into Posts, fetching only a preview of content for performance
// Block fetching is parallelized for better performance when processing multiple pages
func (c *Config) processPages(ctx context.Context, client *notionapi.Client, pages []notionapi.Page) ([]*Post, error) {
	if len(pages) == 0 {
		return []*Post{}, nil
	}

	type pageResult struct {
		post *Post
		err  error
		idx  int
	}

	results := make(chan pageResult, len(pages))

	// Process all pages in parallel
	for i, page := range pages {
		go func(idx int, p notionapi.Page) {
			// Extract tags
			rawTags := p.Properties["Tags"]
			tagData, ok := rawTags.(*notionapi.MultiSelectProperty)
			if !ok {
				results <- pageResult{err: fmt.Errorf("tags property is not a multi-select: %+v", rawTags), idx: idx}
				return
			}
			var tags []string
			for _, tag := range tagData.MultiSelect {
				tags = append(tags, tag.Name)
			}

			// Extract ID
			rawID := p.Properties["ID"]
			idData, ok := rawID.(*notionapi.TitleProperty)
			if !ok {
				results <- pageResult{err: fmt.Errorf("id property is not a title: %+v", rawID), idx: idx}
				return
			}
			id := idData.Title[0].PlainText

			// Fetch only first few blocks for preview (much faster)
			blockResp, err := client.Block.GetChildren(ctx, notionapi.BlockID(p.ID), &notionapi.Pagination{PageSize: 5})
			if err != nil {
				results <- pageResult{err: err, idx: idx}
				return
			}

			text := ""
			for _, block := range blockResp.Results {
				switch block.GetType() {
				case notionapi.BlockTypeParagraph:
					paragraph, ok := block.(*notionapi.ParagraphBlock)
					if !ok {
						results <- pageResult{err: fmt.Errorf("paragraph is incorrect block type: %+v", block), idx: idx}
						return
					}
					text += paragraph.GetRichTextString() + "\n"
				default:
					// Silently skip other block types
				}
			}

			text = strings.TrimSpace(text)

			results <- pageResult{
				post: &Post{
					ID:         id,
					PageID:     p.ID.String(),
					Tags:       tags,
					Text:       text,
					CreatedAt:  p.CreatedTime,
					ModifiedAt: p.LastEditedTime,
				},
				idx: idx,
			}
		}(i, page)
	}

	// Collect results in order
	posts := make([]*Post, len(pages))
	for i := 0; i < len(pages); i++ {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		posts[result.idx] = result.post
	}

	return posts, nil
}

// GetPostFullContent fetches the full content of a post by page ID
func (c *Config) GetPostFullContent(ctx context.Context, pageID string) (string, error) {
	client := c.GetClient()

	// Fetch all blocks directly using page ID
	text := ""
	var cursor string
	for {
		pagination := &notionapi.Pagination{PageSize: 100}
		if cursor != "" {
			pagination.StartCursor = notionapi.Cursor(cursor)
		}

		blockResp, err := client.Block.GetChildren(ctx, notionapi.BlockID(pageID), pagination)
		if err != nil {
			return "", err
		}

		for _, block := range blockResp.Results {
			switch block.GetType() {
			case notionapi.BlockTypeParagraph:
				paragraph, ok := block.(*notionapi.ParagraphBlock)
				if !ok {
					return "", fmt.Errorf("paragraph is incorrect block type: %+v", block)
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

	return strings.TrimSpace(text), nil
}

func (c *Config) getDatabaseID(ctx context.Context) (notionapi.DatabaseID, error) {
	// Return cached database ID if available
	if c.cachedDbID != "" {
		return c.cachedDbID, nil
	}

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

	// Cache the database ID for future use
	c.cachedDbID = id

	return id, nil
}
