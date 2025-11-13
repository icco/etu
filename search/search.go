package search

import (
	"strings"

	"github.com/icco/etu/client"
)

// MatchScore represents how well a post matches the search query
type MatchScore struct {
	Post  *client.Post
	Score int
}

// SimpleSearch performs a lightweight string matching search.
// Returns posts sorted by relevance (best matches first).
func SimpleSearch(query string, posts []*client.Post) []*client.Post {
	if query == "" {
		return posts
	}

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)
	
	var matches []MatchScore
	
	for _, post := range posts {
		score := 0
		searchText := strings.ToLower(post.Text)
		searchTags := strings.ToLower(strings.Join(post.Tags, " "))
		fullText := searchText + " " + searchTags
		
		// Exact phrase match gets highest score
		if strings.Contains(fullText, queryLower) {
			score += 100
		}
		
		// Count word matches
		matchedWords := 0
		for _, word := range queryWords {
			if strings.Contains(fullText, word) {
				matchedWords++
				score += 10
			}
		}
		
		// Bonus for matching all words
		if matchedWords == len(queryWords) {
			score += 50
		}
		
		// Tag matches get extra points
		for _, tag := range post.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				score += 30
			}
		}
		
		// Only include posts with some match
		if score > 0 {
			matches = append(matches, MatchScore{Post: post, Score: score})
		}
	}
	
	// Sort by score (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Score < matches[j].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
	
	result := make([]*client.Post, len(matches))
	for i, match := range matches {
		result[i] = match.Post
	}
	
	return result
}

