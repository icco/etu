package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func searchPosts(cmd *cobra.Command, args []string) error {
	model := newPostListModel(cfg, 50, "Search Results", true)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// If a post was selected, fetch and print full content
	if finalModel.(postListModel).selected != nil {
		selectedPost := finalModel.(postListModel).selected
		// Always fetch full content since we only fetch previews for list
		fullText, err := cfg.GetPostFullContent(cmd.Context(), selectedPost.PageID)
		if err == nil {
			fmt.Println(fullText)
		} else {
			// Fallback to preview text
			fmt.Println(selectedPost.Text)
		}
	}

	return nil
}
