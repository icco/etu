// Package client provides the etu journal backend client: configuration,
// caching, and gRPC calls used by the CLI and TUI.
package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icco/etu-backend/proto"
)

// Post represents a journal entry (display model for TUI/CLI).
type Post struct {
	PageID    string // Note ID for fetching full content
	Tags      []string
	Text      string
	CreatedAt time.Time
	Images    []*proto.NoteImage
	Audios    []*proto.NoteAudio
}

// Config holds the configuration for the client.
type Config struct {
	APIKey     string
	GRPCTarget string
	grpc       *grpcClients
}

// LoadConfig loads configuration from ~/.config/etu/config.json and environment variables.
// Env ETU_API_KEY and ETU_GRPC_TARGET override file values. If no config file exists and
// no API key is set, a config file is created with the correct structure and an empty key.
func LoadConfig() *Config {
	cf, err := loadConfigFromFile()
	if err != nil {
		log.Printf("etu: reading config: %v", err)
	}

	if cf.APIKey == "" {
		cf.APIKey = os.Getenv("ETU_API_KEY")
	}
	if cf.GRPCTarget == "" {
		cf.GRPCTarget = os.Getenv("ETU_GRPC_TARGET")
	}
	// Trim whitespace so pasted keys or env vars with trailing newlines don't break validation.
	return &Config{
		APIKey:     strings.TrimSpace(cf.APIKey),
		GRPCTarget: strings.TrimSpace(cf.GRPCTarget),
	}
}

// Validate checks that the API key is present.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key required: set ETU_API_KEY or add api_key to config file")
	}
	c.warnIfTargetUnresolvable()
	return nil
}

// warnIfTargetUnresolvable prints a stderr warning when GRPCTarget's host doesn't resolve.
// Catches stale grpc_target values after a default change (e.g. PR #97 moved natwelch.com → timeclimbers.com).
func (c *Config) warnIfTargetUnresolvable() {
	if c.GRPCTarget == "" {
		return
	}
	host, _, err := net.SplitHostPort(c.GRPCTarget)
	if err != nil {
		host = c.GRPCTarget
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if addrs, err := net.DefaultResolver.LookupHost(ctx, host); err == nil && len(addrs) > 0 {
		return
	}
	path, _ := ConfigPath()
	fmt.Fprintf(os.Stderr, "etu: warning: grpc_target %q does not resolve (default is %s). Update %s, set ETU_GRPC_TARGET, or clear the value to fall back to the default.\n", c.GRPCTarget, defaultGRPCTarget, path)
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
	cache, err := c.cacheFromFile()
	if err != nil {
		log.Printf("etu: reading timesince cache: %v", err)
	}
	if cache != nil {
		if time.Since(cache.Saved) < 5*time.Minute {
			return cache.Duration, nil
		}
	}
	if err := c.UpdateCache(ctx); err != nil {
		return 0, fmt.Errorf("updating cache %w", err)
	}
	cache, err = c.cacheFromFile()
	if err != nil {
		log.Printf("etu: reading timesince cache: %v", err)
	}
	if cache != nil {
		return cache.Duration, nil
	}
	return 0, fmt.Errorf("cache still not found")
}

type cacheData struct {
	Saved    time.Time
	Duration time.Duration
}

func (c *Config) cachePath() (string, error) {
	return CachePath("timesince.cache")
}

func (c *Config) cacheToFile(dur time.Duration) (err error) {
	path, err := c.cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	// path is built from CachePath() (fixed config dir under user home), not external input.
	f, err := os.Create(path) //nolint:gosec // G304: path is from fixed config dir, not user-controlled
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	return gob.NewEncoder(f).Encode(cacheData{Saved: time.Now(), Duration: dur})
}

func (c *Config) cacheFromFile() (data *cacheData, err error) {
	path, err := c.cachePath()
	if err != nil {
		return nil, err
	}
	// path is built from CachePath() (fixed config dir under user home), not external input.
	f, err := os.Open(path) //nolint:gosec // G304: path is from fixed config dir, not user-controlled
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	data = &cacheData{}
	if err := gob.NewDecoder(f).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

// toInt32 narrows an int to int32, returning an error if it would overflow.
func toInt32(n int) (int32, error) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		return 0, fmt.Errorf("value %d out of int32 range", n)
	}
	return int32(n), nil
}

// detectMIME returns the MIME type of data, falling back to the file extension.
func detectMIME(data []byte, path string) string {
	mimeType := http.DetectContentType(data)
	if mimeType == "application/octet-stream" {
		if ext := filepath.Ext(path); ext != "" {
			if byExt := mime.TypeByExtension(ext); byExt != "" {
				mimeType = byExt
			}
		}
	}
	return mimeType
}

// LoadImageUploads reads image files from paths and returns proto ImageUpload messages.
// MIME type is detected from content (or file extension as fallback).
func LoadImageUploads(paths []string) ([]*proto.ImageUpload, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	out := make([]*proto.ImageUpload, 0, len(paths))
	for _, p := range paths {
		// Paths come from CLI flags supplied by the user; reading them is the intent.
		data, err := os.ReadFile(p) //nolint:gosec // G304: user-supplied CLI input
		if err != nil {
			return nil, fmt.Errorf("read image %s: %w", p, err)
		}
		out = append(out, &proto.ImageUpload{
			Data:     data,
			MimeType: detectMIME(data, p),
		})
	}
	return out, nil
}

// LoadAudioUploads reads audio files from paths and returns proto AudioUpload messages.
// MIME type is detected from content (or file extension as fallback).
func LoadAudioUploads(paths []string) ([]*proto.AudioUpload, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	out := make([]*proto.AudioUpload, 0, len(paths))
	for _, p := range paths {
		// Paths come from CLI flags supplied by the user; reading them is the intent.
		data, err := os.ReadFile(p) //nolint:gosec // G304: user-supplied CLI input
		if err != nil {
			return nil, fmt.Errorf("read audio %s: %w", p, err)
		}
		out = append(out, &proto.AudioUpload{
			Data:     data,
			MimeType: detectMIME(data, p),
		})
	}
	return out, nil
}

// SaveEntry saves a new journal entry via the backend (tags are generated on the backend).
// imagePaths and audioPaths are optional paths to image and audio files to attach to the note.
func (c *Config) SaveEntry(ctx context.Context, text string, imagePaths, audioPaths []string) error {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return err
	}
	images, err := LoadImageUploads(imagePaths)
	if err != nil {
		return err
	}
	audios, err := LoadAudioUploads(audioPaths)
	if err != nil {
		return err
	}
	resp, err := g.notesClient.CreateNote(ctx, &proto.CreateNoteRequest{
		UserId:  userID,
		Content: text,
		Images:  images,
		Audios:  audios,
	})
	if err != nil {
		return err
	}
	created := resp.GetNote()
	if created != nil && created.GetCreatedAt() != nil {
		dur := time.Since(created.GetCreatedAt().AsTime())
		if err := c.cacheToFile(dur); err != nil {
			log.Printf("etu: updating timesince cache: %v", err)
		}
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
	limit, err := toInt32(count)
	if err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}
	resp, err := g.notesClient.ListNotes(ctx, &proto.ListNotesRequest{
		UserId: userID,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return notesToPosts(resp.GetNotes()), nil
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
	limit, err := toInt32(maxResults)
	if err != nil {
		return nil, fmt.Errorf("maxResults: %w", err)
	}
	resp, err := g.notesClient.ListNotes(ctx, &proto.ListNotesRequest{
		UserId: userID,
		Search: query,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return notesToPosts(resp.GetNotes()), nil
}

// GetRandomPosts fetches random journal entries from the backend.
func (c *Config) GetRandomPosts(ctx context.Context, count int) ([]*Post, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return nil, err
	}
	c32, err := toInt32(count)
	if err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}
	resp, err := g.notesClient.GetRandomNotes(ctx, &proto.GetRandomNotesRequest{
		UserId: userID,
		Count:  c32,
	})
	if err != nil {
		return nil, err
	}
	return notesToPosts(resp.GetNotes()), nil
}

// Tag represents a journal tag and how many entries use it.
type Tag struct {
	Name  string
	Count int32
}

// ListTags lists all tags for the current user.
func (c *Config) ListTags(ctx context.Context) ([]Tag, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return nil, err
	}
	resp, err := g.tagsClient.ListTags(ctx, &proto.ListTagsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return protoTagsToTags(resp.GetTags()), nil
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
