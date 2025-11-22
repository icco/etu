package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func searchPosts(cmd *cobra.Command, args []string) error {
	var query string

	// Use huh for search input
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Search journal entries").
				Value(&query).
				Placeholder("Enter search query..."),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		// If empty query, just list recent posts
		return listPosts(cmd, args)
	}

	// Run the list model in search mode with the query
	model := newPostListModel(cfg, 50, "Search Results", false)
	model.query = query
	model.loading = true

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// If a post was selected, fetch and print full content
	if finalModel.(postListModel).selected != nil {
		selectedPost := finalModel.(postListModel).selected

		// Fetch with spinner
		type fetchResult struct {
			text string
			err  error
		}
		resultChan := make(chan interface{}, 1)

		go func() {
			fullText, err := cfg.GetPostFullContent(cmd.Context(), selectedPost.PageID)
			resultChan <- fetchResult{text: fullText, err: err}
		}()

		spinnerModel := newSpinnerModel("Loading full content...", resultChan)
		p := tea.NewProgram(spinnerModel)
		if _, err := p.Run(); err != nil {
			return err
		}

		result := (<-resultChan).(fetchResult)
		if result.err == nil {
			fmt.Println(result.text)
		} else {
			// Fallback to preview text
			fmt.Println(selectedPost.Text)
		}
	}

	return nil
}
