package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var (
	Version   = ""
	CommitSHA = ""

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
		Use:     "last",
		Aliases: []string{"l"},
		Short:   "Output a string of time since last post.",
		Args:    cobra.NoArgs,
		RunE:    mostRecentPost,
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
		Aliases: []string{"ls"},
		Short:   "List journal entries, with an optional starting datetime.",
		Args:    cobra.NoArgs,
		RunE:    listPosts,
	}

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync local db with cloud db.",
		Args:  cobra.NoArgs,
		RunE:  syncPosts,
	}
)

func createPost(cmd *cobra.Command, args []string) error {
	model := client.CreateModel()
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		return err
	}

	return client.SaveEntry(cmd.Context(), string(model.Data))
}

func timeSinceLastPost(cmd *cobra.Command, args []string) error {
	dur, err := client.TimeSinceLastPost(cmd.Context())
	if err != nil {
		return err
	}

	fmt.Printf("%0.1fh", dur.Hours())

	return nil
}

func deletePost(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("delete takes only one argument")
	}

	return client.DeletePost(cmd.Context(), args[0])
}

func renderPosts(entries []*client.Post) error {
	for _, e := range entries {
		in := fmt.Sprintf("# %s\n%s\n", e.CreatedAt, e.Content)

		r, _ := glamour.NewTermRenderer(
			// detect background color and pick either the default dark or light theme
			glamour.WithAutoStyle(),
			// wrap output at specific width
			glamour.WithWordWrap(80),
		)

		out, err := r.Render(in)
		if err != nil {
			return err
		}

		fmt.Print(out)
	}

	return nil

}

func mostRecentPost(cmd *cobra.Command, args []string) error {
	entries, err := client.ListPosts(cmd.Context(), 1)
	if err != nil {
		return err
	}

	return renderPosts(entries)
}

func listPosts(cmd *cobra.Command, args []string) error {
	entries, err := client.ListPosts(cmd.Context(), 25)
	if err != nil {
		return err
	}

	return renderPosts(entries)
}

func syncPosts(cmd *cobra.Command, args []string) error {
	return client.Sync(cmd.Context())
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
		syncCmd,
		timeSinceCmd,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
