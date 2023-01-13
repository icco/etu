package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFile = "etu.db"
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

	timeSinceCmd = &cobra.Command{
		Use:     "timesince",
		Aliases: []string{"tslp"},
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
	db, err := openKV()
	if err != nil {
		return err
	}

	model := client.CreateModel()
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		return err
	}

	return client.SaveEntry(cmd.Context(), db, time.Now(), string(model.Data))
}

func timeSinceLastPost(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	dur, err := client.TimeSinceLastPost(cmd.Context(), db)
	if err != nil {
		return err
	}

	fmt.Printf("%s", dur.String())

	return nil
}

func deletePost(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("delete takes only one argument")
	}

	return client.DeleteEntry(cmd.Context(), db, args[0])
}

func listPosts(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	entries, err := client.ListEntries(cmd.Context(), db, 10)
	if err != nil {
		return err
	}

	for _, e := range entries {
		in := fmt.Sprintf("# %s\n%s\n", e.Key, e.Data)

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

func syncPosts(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	return client.Sync(db)
}

func openKV() (*sql.DB, error) {
	return sql.Open("sqlite3", dbFile)
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
