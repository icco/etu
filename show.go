package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a journal entry with images and audio.",
	Args:  cobra.NoArgs,
	RunE:  showPost,
}

func showPost(cmd *cobra.Command, args []string) error {
	// Show list of posts to select from
	model := newPostListModel(cfg, 25, "Select entry to view", true)
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
	return displayPost(cmd, selectedPost)
}

func displayPost(cmd *cobra.Command, post *client.Post) error {
	// Fetch full content
	var fullText string
	var fetchErr error
	err := spinner.New().
		Title("Loading full content...").
		Action(func() {
			fullText, fetchErr = cfg.GetPostFullContent(cmd.Context(), post.PageID)
		}).
		Run()

	if err != nil {
		return err
	}
	if fetchErr != nil {
		fullText = post.Text
	}

	// Display header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println()
	fmt.Println(headerStyle.Render("Date: ") + post.CreatedAt.Format("2006-01-02 15:04"))

	if len(post.Tags) > 0 {
		fmt.Println(headerStyle.Render("Tags: ") + strings.Join(post.Tags, ", "))
	}

	fmt.Println()
	fmt.Println(fullText)

	// Display images
	if len(post.Images) > 0 {
		fmt.Println()
		fmt.Println(headerStyle.Render("Images:"))
		for i, img := range post.Images {
			fmt.Printf("  %d. %s\n", i+1, img.URL)
			if img.ExtractedText != "" {
				fmt.Println(labelStyle.Render("     Text: ") + truncate(img.ExtractedText, 80))
			}
			// Try to display image inline if terminal supports it
			displayImageInline(img.URL)
		}
	}

	// Display audio
	if len(post.Audios) > 0 {
		fmt.Println()
		fmt.Println(headerStyle.Render("Audio:"))
		for i, aud := range post.Audios {
			fmt.Printf("  %d. %s\n", i+1, aud.URL)
			if aud.TranscribedText != "" {
				fmt.Println(labelStyle.Render("     Transcription: ") + truncate(aud.TranscribedText, 80))
			}
		}
	}

	fmt.Println()
	return nil
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// displayImageInline attempts to display an image inline using iTerm2's protocol
func displayImageInline(url string) {
	// Check if we're in iTerm2
	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		return
	}

	// Fetch the image
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// iTerm2 inline image protocol
	encoded := base64.StdEncoding.EncodeToString(data)
	fmt.Printf("\033]1337;File=inline=1;width=auto;height=10;preserveAspectRatio=1:%s\007\n", encoded)
}
