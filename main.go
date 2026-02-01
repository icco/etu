package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
			// Skip API key validation for these commands (they don't need the backend)
			curr := cmd
			for curr != nil {
				if curr.Name() == "completion" || curr.Name() == "help" || curr.Name() == "__complete" {
					return nil
				}
				curr = curr.Parent()
			}

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
		Short:   "Create a new journal entry (attach images with -i, audio with -a, or in TUI).",
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
	var imagePathsInput string
	var audioPathsInput string

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin is a pipe or redirected input
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		text = string(content)
	} else {
		// stdin is a terminal, use interactive TUI (supports drag & drop of images)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Value(&text).
					Placeholder("Write your journal entry here...").
					Validate(func(value string) error {
						if len(strings.TrimSpace(value)) == 0 {
							return fmt.Errorf("journal entry cannot be empty")
						}
						return nil
					}).
					WithHeight(12).
					WithWidth(100),
				huh.NewText().
					Value(&imagePathsInput).
					Title("Images").
					Description("Drag & drop image files here, or paste paths (one per line). Leave empty for no images.").
					Placeholder("/path/to/image.jpg").
					WithHeight(3).
					WithWidth(100),
				huh.NewText().
					Value(&audioPathsInput).
					Title("Audio").
					Description("Drag & drop audio files here, or paste paths (one per line). Leave empty for no audio.").
					Placeholder("/path/to/recording.mp3").
					WithHeight(3).
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

	imagePaths, err := cmd.Flags().GetStringSlice("image")
	if err != nil {
		imagePaths = nil
	}
	audioPaths, err := cmd.Flags().GetStringSlice("audio")
	if err != nil {
		audioPaths = nil
	}
	// Parse TUI image paths (from drag & drop or paste): one path per line, trim spaces
	if imagePathsInput != "" {
		for _, line := range strings.Split(imagePathsInput, "\n") {
			p := strings.TrimSpace(line)
			if p == "" {
				continue
			}
			// Strip quotes terminals sometimes add around paths with spaces
			p = strings.Trim(p, `"'`)
			if abs, err := filepath.Abs(p); err == nil {
				p = abs
			}
			imagePaths = append(imagePaths, p)
		}
	}
	// Parse TUI audio paths (from drag & drop or paste): one path per line, trim spaces
	if audioPathsInput != "" {
		for _, line := range strings.Split(audioPathsInput, "\n") {
			p := strings.TrimSpace(line)
			if p == "" {
				continue
			}
			p = strings.Trim(p, `"'`)
			if abs, err := filepath.Abs(p); err == nil {
				p = abs
			}
			audioPaths = append(audioPaths, p)
		}
	}

	// Save entry with spinner
	var saveErr error
	err = spinner.New().
		Title("Saving entry...").
		Action(func() {
			saveErr = cfg.SaveEntry(cmd.Context(), text, imagePaths, audioPaths)
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
			deleteErr = cfg.DeletePost(cmd.Context(), selectedPost.PageID)
		}).
		Run()

	if err != nil {
		return err
	}

	return deleteErr
}

func mostRecentPost(cmd *cobra.Command, args []string) error {
	// Fetch most recent post (preview first)
	posts, err := cfg.ListPosts(cmd.Context(), 1)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return fmt.Errorf("no posts found")
	}
	post := posts[0]

	// Detect if stdout is a terminal (interactive) or being piped.
	stdoutStat, err := os.Stdout.Stat()
	interactive := err == nil && (stdoutStat.Mode()&os.ModeCharDevice) != 0

	if !interactive {
		// Non-interactive: output full content for piping.
		full, fullErr := cfg.GetPostFullContent(cmd.Context(), post.PageID)
		if fullErr == nil && strings.TrimSpace(full) != "" {
			fmt.Print(full)
			return nil
		}
		// Fallback to preview text if full fetch fails.
		fmt.Print(post.Text)
		return nil
	}

	// Interactive: show existing TUI list (single item) for consistency.
	model := newPostListModel(cfg, 1, "Most Recent Entry", true)
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

	createCmd.Flags().StringSliceP("image", "i", nil, "path to image file to attach (can be repeated)")
	createCmd.Flags().StringSliceP("audio", "a", nil, "path to audio file to attach (can be repeated)")

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
	// Load config (file + ETU_API_KEY / ETU_GRPC_TARGET env) and persist so we don't have to mess with env later.
	cfg = client.LoadConfig()
	if _, err := client.SaveConfig(cfg.ApiKey, cfg.GRPCTarget); err != nil {
		log.Fatal(err)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
