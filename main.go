package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the application.
	Version = ""
	// CommitSHA is the git commit SHA of the build.
	CommitSHA = ""

	cfg *client.Config

	rootCmd = &cobra.Command{
		Use:   "etu",
		Short: "Etu. A personal command line journal.",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config initialization for these commands
			curr := cmd
			for curr != nil {
				if curr.Name() == "completion" || curr.Name() == "help" || curr.Name() == "__complete" {
					return nil
				}
				curr = curr.Parent()
			}

			cfg = client.LoadConfig()
			if err := cfg.Validate(); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	createCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "new"},
		Short:   "Create a new journal entry. If no date provided, current time will be used.",
		Args:    cobra.NoArgs,
		RunE:    createPost,
	}

	deleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d"},
		Short:   "Delete a journal entry.",
		Args:    cobra.NoArgs,
		RunE:    deletePost,
	}

	mostRecentCmd = &cobra.Command{
		Use:   "last",
		Short: "Output the most recent journal entry.",
		Args:  cobra.NoArgs,
		RunE:  mostRecentPost,
	}

	timeSinceCmd = &cobra.Command{
		Use:     "timesince",
		Aliases: []string{"ts", "tslp"},
		Short:   "Output a string of time since last post.",
		Args:    cobra.NoArgs,
		RunE:    timeSinceLastPost,
	}

	listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List journal entries, with an optional starting datetime.",
		Args:    cobra.NoArgs,
		RunE:    listPosts,
	}

	searchCmd = &cobra.Command{
		Use:     "search",
		Aliases: []string{"s"},
		Short:   "Search journal entries using fuzzy search.",
		Args:    cobra.NoArgs,
		RunE:    searchPosts,
	}
)

func createPost(cmd *cobra.Command, args []string) error {
	// Check if stdin has data (piped input)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	var text string

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin is a pipe or redirected input
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		text = string(content)
	} else {
		// stdin is a terminal, use interactive TUI
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Value(&text).
					Placeholder("Write your journal entry here...").
					Validate(func(value string) error {
						if len(value) == 0 {
							return fmt.Errorf("journal entry cannot be empty")
						}
						return nil
					}).
					WithHeight(15).
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
	}

	// Save entry with spinner
	var saveErr error
	err = spinner.New().
		Title("Saving entry...").
		Action(func() {
			saveErr = cfg.SaveEntry(cmd.Context(), text)
		}).
		Run()

	if err != nil {
		return err
	}

	return saveErr
}

func timeSinceLastPost(cmd *cobra.Command, args []string) error {
	ret := "???"
	dur, err := cfg.TimeSinceLastPost(cmd.Context())
	if err == nil {

		switch {
		case dur.Hours() > 24:
			ret = fmt.Sprintf("%0.1fd", dur.Hours()/24)
		default:
			ret = fmt.Sprintf("%0.1fh", dur.Hours())
		}
	}

	fmt.Print(ret)

	return nil
}

func deletePost(cmd *cobra.Command, args []string) error {
	// Show list of posts to select from
	model := newPostListModel(cfg, 25, "Select entry to delete", true)
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

	// Prompt for confirmation
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete entry from %s?", selectedPost.CreatedAt.Format("2006-01-02 15:04"))).
				Description(selectedPost.Text).
				Value(&confirm),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return err
	}

	if !confirm {
		return nil // User chose not to delete
	}

	// Delete with spinner
	var deleteErr error
	err = spinner.New().
		Title("Deleting entry...").
		Action(func() {
			deleteErr = cfg.DeletePost(cmd.Context(), selectedPost.ID)
		}).
		Run()

	if err != nil {
		return err
	}

	return deleteErr
}

func mostRecentPost(cmd *cobra.Command, args []string) error {
	model := newPostListModel(cfg, 1, "Interstitial Notes", true)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return err
	}
	return nil
}

func listPosts(cmd *cobra.Command, args []string) error {
	model := newPostListModel(cfg, 25, "Interstitial Notes", true)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	if len(CommitSHA) >= 7 {
		vt := rootCmd.VersionTemplate()
		rootCmd.SetVersionTemplate(vt[:len(vt)-1] + " (" + CommitSHA[0:7] + ")\n")
	}
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(
		createCmd,
		deleteCmd,
		listCmd,
		mostRecentCmd,
		timeSinceCmd,
		searchCmd,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
