package main

import (
	"fmt"
	"math"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var (
	Version   = ""
	CommitSHA = ""

	cfg *client.Config

	rootCmd = &cobra.Command{
		Use:   "etu",
		Short: "Etu. A personal command line journal.",
		Args:  cobra.NoArgs,
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
		Use:     "delete ID",
		Aliases: []string{"d"},
		Short:   "Delete a journal entry.",
		Args:    cobra.ExactArgs(1),
		RunE:    deletePost,
	}

	mostRecentCmd = &cobra.Command{
		Use:   "last",
		Short: "Output a string of time since last post.",
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
	model := createModel()
	p := tea.NewProgram(model)
	_, err := p.Run()
	if err != nil {
		return err
	}

	return cfg.SaveEntry(cmd.Context(), string(model.Data))
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
	if len(args) != 1 {
		return fmt.Errorf("delete takes only one argument")
	}

	return cfg.DeletePost(cmd.Context(), args[0])
}

func renderPosts(entries []*client.Post) error {
	var items []list.Item
	for _, e := range entries {
		items = append(items, listItem{post: e})
	}

	buffer := 6
	maxSize := 10
	height := math.Min(float64(maxSize+buffer), float64(len(items)+buffer))

	m := listModel{list: list.New(items, itemDelegate{}, 0, int(height))}
	m.list.Title = "Interstitial Notes"
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowTitle(true)
	m.list.SetShowHelp(true)

	m.list.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	return nil

}

func mostRecentPost(cmd *cobra.Command, args []string) error {
	entries, err := cfg.ListPosts(cmd.Context(), 1)
	if err != nil {
		return err
	}

	return renderPosts(entries)
}

func listPosts(cmd *cobra.Command, args []string) error {
	entries, err := cfg.ListPosts(cmd.Context(), 25)
	if err != nil {
		return err
	}

	return renderPosts(entries)
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
	var err error
	cfg, err = client.New(os.Getenv("NOTION_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
