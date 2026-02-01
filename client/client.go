package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icco/etu-backend/proto"
)

// Post represents a journal entry (display model for TUI/CLI).
type Post struct {
	ID        string
	PageID    string    // Note ID for fetching full content
	Tags      []string
	Text      string
	CreatedAt time.Time
	ModifiedAt time.Time
}

// Config holds the configuration for the client.
type Config struct {
	ApiKey     string
	GRPCTarget string
	grpc       *grpcClients
}

// LoadConfig loads configuration from ~/.config/etu/config.json and environment variables.
// Env ETU_API_KEY and ETU_GRPC_TARGET override file values.
func LoadConfig() *Config {
	apiKey, grpcTarget, _ := loadConfigFromFile()
	if apiKey == "" {
		apiKey = os.Getenv("ETU_API_KEY")
	}
	if grpcTarget == "" {
		grpcTarget = os.Getenv("ETU_GRPC_TARGET")
	}
	if grpcTarget == "" {
		grpcTarget = defaultGRPCTarget
	}
	return &Config{
		ApiKey:     apiKey,
		GRPCTarget: grpcTarget,
	}
}

// Validate checks that the API key is present.
func (c *Config) Validate() error {
	if c.ApiKey == "" {
		dir, _ := ConfigDir()
		if dir != "" {
			return fmt.Errorf("API key required: set ETU_API_KEY or add api_key to %s/config.json (see https://github.com/icco/etu-backend)", dir)
		}
		return fmt.Errorf("API key required: set ETU_API_KEY or add api_key to ~/.config/etu/config.json (see https://github.com/icco/etu-backend)")
	}
	return nil
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
	return c.cacheToFile(dur)
}

// TimeSinceLastPost returns the time since the last post was created.
func (c *Config) TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	cache, _ := c.cacheFromFile()
	if cache != nil {
		if time.Since(cache.Saved) < 5*time.Minute {
			return cache.Duration, nil
		}
	}
	if err := c.UpdateCache(ctx); err != nil {
		return 0, fmt.Errorf("updating cache %w", err)
	}
	if cache, _ := c.cacheFromFile(); cache != nil {
		return cache.Duration, nil
	}
	return 0, fmt.Errorf("cache still not found")
}

type cacheData struct {
	Saved    time.Time
	Duration time.Duration
}

func (c *Config) cachePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "timesince.cache"), nil
}

func (c *Config) cacheToFile(dur time.Duration) error {
	path, err := c.cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(cacheData{Saved: time.Now(), Duration: dur})
}

func (c *Config) cacheFromFile() (*cacheData, error) {
	path, err := c.cachePath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data := &cacheData{}
	if err := gob.NewDecoder(f).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

// SaveEntry saves a new journal entry via the backend (tags are generated on the backend).
func (c *Config) SaveEntry(ctx context.Context, text string) error {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return err
	}
	resp, err := g.notesClient.CreateNote(ctx, &proto.CreateNoteRequest{
		UserId:  userID,
		Content: text,
	})
	if err != nil {
		return err
	}
	created := resp.GetNote()
	if created != nil && created.GetCreatedAt() != nil {
		dur := time.Since(created.GetCreatedAt().AsTime())
		_ = c.cacheToFile(dur)
	}
	return nil
}

// DeletePost deletes a journal entry by ID.
func (c *Config) DeletePost(ctx context.Context, pageID string) error {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return err
	}
	_, err = g.notesClient.DeleteNote(ctx, &proto.DeleteNoteRequest{
		UserId: userID,
		Id:     pageID,
	})
	return err
}

// ListPosts lists the most recent journal entries.
func (c *Config) ListPosts(ctx context.Context, count int) ([]*Post, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return nil, err
	}
	resp, err := g.notesClient.ListNotes(ctx, &proto.ListNotesRequest{
		UserId: userID,
		Limit:  int32(count),
	})
	if err != nil {
		return nil, err
	}
	posts := make([]*Post, 0, len(resp.GetNotes()))
	for _, n := range resp.GetNotes() {
		posts = append(posts, noteToPost(n))
	}
	return posts, nil
}

// SearchPosts searches journal entries via the backend.
func (c *Config) SearchPosts(ctx context.Context, query string, maxResults int) ([]*Post, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return nil, err
	}
	resp, err := g.notesClient.ListNotes(ctx, &proto.ListNotesRequest{
		UserId: userID,
		Search: query,
		Limit:  int32(maxResults),
	})
	if err != nil {
		return nil, err
	}
	posts := make([]*Post, 0, len(resp.GetNotes()))
	for _, n := range resp.GetNotes() {
		posts = append(posts, noteToPost(n))
	}
	return posts, nil
}

// GetPostFullContent fetches the full content of a post by ID.
func (c *Config) GetPostFullContent(ctx context.Context, pageID string) (string, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return "", err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return "", err
	}
	resp, err := g.notesClient.GetNote(ctx, &proto.GetNoteRequest{
		UserId: userID,
		Id:     pageID,
	})
	if err != nil {
		return "", err
	}
	if n := resp.GetNote(); n != nil {
		return strings.TrimSpace(n.GetContent()), nil
	}
	return "", nil
}
