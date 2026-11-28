package main

import (
	"strings"

	"charm.land/huh/v2"
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

	if err := form.WithTheme(etuTheme).Run(); err != nil {
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

	selected, err := runPostList(model)
	if err != nil || selected == nil {
		return err
	}
	return displayPost(cmd, selected)
}
