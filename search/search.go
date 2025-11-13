package search

import (
	"strings"

	"github.com/icco/etu/client"
	"github.com/sahilm/fuzzy"
)

// SearchPosts performs fuzzy search on journal entries.
// It searches through the text content and tags of each post.
// Returns posts sorted by relevance (best matches first).
func SearchPosts(query string, posts []*client.Post) []*client.Post {
	if query == "" {
		return posts
	}

	// Create searchable strings from posts
	type searchablePost struct {
		post   *client.Post
		search string
	}
	var searchable []searchablePost
	var searchStrings []string

	for _, p := range posts {
		searchStr := strings.ToLower(p.Text)
		// Also include tags in search
		if len(p.Tags) > 0 {
			searchStr += " " + strings.ToLower(strings.Join(p.Tags, " "))
		}
		searchable = append(searchable, searchablePost{
			post:   p,
			search: searchStr,
		})
		searchStrings = append(searchStrings, searchStr)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(strings.ToLower(query), searchStrings)

	// Map results back to posts
	result := make([]*client.Post, len(matches))
	for i, match := range matches {
		result[i] = searchable[match.Index].post
	}

	return result
}

