package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit a journal entry.",
	Args:    cobra.NoArgs,
	RunE:    editPost,
}

func editPost(cmd *cobra.Command, _ []string) error {
	// Show list of posts to select from
	model := newPostListModel(cfg, 25, "Select entry to edit", true)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if a post was selected
	if finalModel.(postListModel).selected == nil {
		return nil // User quit without selecting
	}

	selectedPost := finalModel.(postListModel).selected

	// Fetch full content so the editor is pre-filled with the whole entry
	var text string
	var fetchErr error
	err = spinner.New().
		Title("Loading full content...").
		Action(func() {
			text, fetchErr = cfg.GetPostFullContent(cmd.Context(), selectedPost.PageID)
		}).
		Run()

	if err != nil {
		return err
	}
	if fetchErr != nil {
		return fetchErr
	}
	original := text

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Value(&text).
				Title(fmt.Sprintf("Edit entry from %s", selectedPost.CreatedAt.Format("2006-01-02 15:04"))).
				Validate(func(value string) error {
					if len(strings.TrimSpace(value)) == 0 {
						return fmt.Errorf("journal entry cannot be empty")
					}
					return nil
				}).
				WithHeight(12).
				WithWidth(100),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("journal entry cannot be empty")
	}
	if text == strings.TrimSpace(original) {
		fmt.Println("No changes.")
		return nil
	}

	// Update with spinner
	var updateErr error
	err = spinner.New().
		Title("Updating entry...").
		Action(func() {
			_, updateErr = cfg.UpdatePost(cmd.Context(), selectedPost.PageID, text)
		}).
		Run()

	if err != nil {
		return err
	}

	return updateErr
}
