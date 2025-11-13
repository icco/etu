package search

import (
	"strings"

	"github.com/icco/etu/client"
	"github.com/sahilm/fuzzy"
)

// SearchablePosts wraps posts with pre-computed search strings for efficient searching.
type SearchablePosts struct {
	posts        []*client.Post
	searchStrings []string
}

// NewSearchablePosts creates a new SearchablePosts instance with pre-computed search strings.
func NewSearchablePosts(posts []*client.Post) *SearchablePosts {
	searchStrings := make([]string, len(posts))
	for i, p := range posts {
		searchStr := strings.ToLower(p.Text)
		// Also include tags in search
		if len(p.Tags) > 0 {
			searchStr += " " + strings.ToLower(strings.Join(p.Tags, " "))
		}
		searchStrings[i] = searchStr
	}

	return &SearchablePosts{
		posts:         posts,
		searchStrings: searchStrings,
	}
}

// Search performs fuzzy search on the pre-computed search strings.
// Returns posts sorted by relevance (best matches first).
func (sp *SearchablePosts) Search(query string) []*client.Post {
	if query == "" {
		return sp.posts
	}

	// Perform fuzzy search
	matches := fuzzy.Find(strings.ToLower(query), sp.searchStrings)

	// Map results back to posts
	result := make([]*client.Post, len(matches))
	for i, match := range matches {
		result[i] = sp.posts[match.Index]
	}

	return result
}

