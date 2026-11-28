// Package client provides the etu journal backend client: configuration,
// caching, and gRPC calls used by the CLI and TUI.
package client

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.icco.me/etu-backend/proto"
)

const (
	timesinceCacheFile = "timesince.cache"
	timesinceCacheTTL  = 5 * time.Minute
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
// Env ETU_API_KEY and ETU_GRPC_TARGET override file values. It never writes the config file.
func LoadConfig() *Config {
	cf, err := loadConfigFromFile()
	if err != nil {
		log.Printf("etu: reading config: %v", err)
		cf = &ConfigFile{}
	}

	if v := os.Getenv("ETU_API_KEY"); v != "" {
		cf.APIKey = v
	}
	if v := os.Getenv("ETU_GRPC_TARGET"); v != "" {
		cf.GRPCTarget = v
	}
	if strings.TrimSpace(cf.GRPCTarget) == "" {
		cf.GRPCTarget = defaultGRPCTarget
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
		return fmt.Errorf("API key required: run `etu login` or set ETU_API_KEY (sign up at %s)", SignupURL)
	}
	return nil
}

// UpdateCache updates the cache with the latest post.
func (c *Config) UpdateCache(ctx context.Context) error {
	last, err := c.lastPostTime(ctx)
	if err != nil {
		return err
	}
	return c.cacheToFile(last)
}

// TimeSinceLastPost returns the time since the last post was created.
func (c *Config) TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	path, err := c.cachePath()
	if err != nil {
		return 0, err
	}
	return timeSinceLastPost(path, func() (time.Time, error) { return c.lastPostTime(ctx) })
}

// timeSinceLastPost answers from the cache at path while it is fresh, otherwise
// from fetch. A failed cache write is logged, since the answer is still right.
func timeSinceLastPost(path string, fetch func() (time.Time, error)) (time.Duration, error) {
	if cache, err := readCache(path); err == nil && cache != nil && !cache.LastPost.IsZero() && time.Since(cache.Saved) < timesinceCacheTTL {
		return time.Since(cache.LastPost), nil
	}
	last, err := fetch()
	if err != nil {
		return 0, err
	}
	if err := writeCache(path, last); err != nil {
		log.Printf("etu: writing timesince cache: %v", err)
	}
	return time.Since(last), nil
}

func (c *Config) lastPostTime(ctx context.Context) (time.Time, error) {
	posts, err := c.ListPosts(ctx, 1)
	if err != nil {
		return time.Time{}, err
	}
	if len(posts) == 0 {
		return time.Time{}, errors.New("no posts found")
	}
	return posts[0].CreatedAt, nil
}

type cacheData struct {
	Saved    time.Time
	LastPost time.Time
}

func (c *Config) cachePath() (string, error) {
	return CachePath(timesinceCacheFile)
}

func (c *Config) cacheToFile(lastPost time.Time) error {
	path, err := c.cachePath()
	if err != nil {
		return err
	}
	return writeCache(path, lastPost)
}

func (c *Config) cacheFromFile() (*cacheData, error) {
	path, err := c.cachePath()
	if err != nil {
		return nil, err
	}
	return readCache(path)
}

// writeCache replaces the cache file with a rename, so a concurrent reader
// sees the old file or the new one, never a half-written one.
func writeCache(path string, lastPost time.Time) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(f.Name()))
		}
	}()
	if err := gob.NewEncoder(f).Encode(cacheData{Saved: time.Now(), LastPost: lastPost}); err != nil {
		return errors.Join(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func readCache(path string) (data *cacheData, err error) {
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

// loadUploads reads each path and builds an upload from its data and MIME type.
func loadUploads[T any](kind string, paths []string, build func(data []byte, mimeType string) T) ([]T, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	out := make([]T, 0, len(paths))
	for _, p := range paths {
		// Paths come from CLI flags supplied by the user; reading them is the intent.
		data, err := os.ReadFile(p) //nolint:gosec // G304: user-supplied CLI input
		if err != nil {
			return nil, fmt.Errorf("read %s %s: %w", kind, p, err)
		}
		out = append(out, build(data, detectMIME(data, p)))
	}
	return out, nil
}

// LoadImageUploads reads image files from paths and returns proto ImageUpload messages.
// MIME type is detected from content (or file extension as fallback).
func LoadImageUploads(paths []string) ([]*proto.ImageUpload, error) {
	return loadUploads("image", paths, func(data []byte, mimeType string) *proto.ImageUpload {
		return &proto.ImageUpload{Data: data, MimeType: mimeType}
	})
}

// LoadAudioUploads reads audio files from paths and returns proto AudioUpload messages.
// MIME type is detected from content (or file extension as fallback).
func LoadAudioUploads(paths []string) ([]*proto.AudioUpload, error) {
	return loadUploads("audio", paths, func(data []byte, mimeType string) *proto.AudioUpload {
		return &proto.AudioUpload{Data: data, MimeType: mimeType}
	})
}

// SaveEntry saves a new journal entry via the backend (tags are generated on the backend).
// imagePaths and audioPaths are optional paths to image and audio files to attach to the note.
func (c *Config) SaveEntry(ctx context.Context, text string, imagePaths, audioPaths []string) error {
	images, err := LoadImageUploads(imagePaths)
	if err != nil {
		return err
	}
	audios, err := LoadAudioUploads(audioPaths)
	if err != nil {
		return err
	}
	g, userID, err := c.ensureUserID(ctx)
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
	if created := resp.GetNote(); created != nil && created.GetCreatedAt() != nil {
		if err := c.cacheToFile(created.GetCreatedAt().AsTime()); err != nil {
			log.Printf("etu: updating timesince cache: %v", err)
		}
	}
	return nil
}

// UpdatePost updates the content of an existing journal entry by ID.
// Tags are left untouched (the backend keeps the existing tag list).
func (c *Config) UpdatePost(ctx context.Context, pageID, content string) (*Post, error) {
	g, userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := g.notesClient.UpdateNote(ctx, &proto.UpdateNoteRequest{
		UserId:     userID,
		Id:         pageID,
		Content:    &content,
		UpdateTags: false,
	})
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	return noteToPost(resp.GetNote()), nil
}

// DeletePost deletes a journal entry by ID.
func (c *Config) DeletePost(ctx context.Context, pageID string) error {
	g, userID, err := c.ensureUserID(ctx)
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
	return c.SearchPosts(ctx, "", count)
}

// SearchPosts searches journal entries via the backend. An empty query lists
// the most recent entries.
func (c *Config) SearchPosts(ctx context.Context, query string, maxResults int) ([]*Post, error) {
	limit, err := toInt32(maxResults)
	if err != nil {
		return nil, fmt.Errorf("maxResults: %w", err)
	}
	g, userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
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
	c32, err := toInt32(count)
	if err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}
	g, userID, err := c.ensureUserID(ctx)
	if err != nil {
		return nil, err
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
	g, userID, err := c.ensureUserID(ctx)
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

// Stats holds aggregate journal metrics from the backend.
type Stats struct {
	TotalBlips   int64
	UniqueTags   int64
	WordsWritten int64
}

// GetStats fetches aggregate stats for the current user.
func (c *Config) GetStats(ctx context.Context) (Stats, error) {
	g, userID, err := c.ensureUserID(ctx)
	if err != nil {
		return Stats{}, err
	}
	resp, err := g.statsClient.GetStats(ctx, &proto.GetStatsRequest{
		UserId: userID,
	})
	if err != nil {
		return Stats{}, fmt.Errorf("get stats: %w", err)
	}
	return Stats{
		TotalBlips:   resp.GetTotalBlips(),
		UniqueTags:   resp.GetUniqueTags(),
		WordsWritten: resp.GetWordsWritten(),
	}, nil
}

// GetPostFullContent fetches the full content of a post by ID.
func (c *Config) GetPostFullContent(ctx context.Context, pageID string) (string, error) {
	g, userID, err := c.ensureUserID(ctx)
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
