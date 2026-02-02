package main

import (
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

	// If a post was selected, display it with media
	if finalModel.(postListModel).selected != nil {
		return displayPost(cmd, finalModel.(postListModel).selected)
	}

	return nil
}
