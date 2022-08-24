package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/charm/cmd"
	"github.com/charmbracelet/charm/kv"
	"github.com/charmbracelet/glamour"
	"github.com/dgraph-io/badger/v3"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

const (
	dbName = "charm.sh.etu.default"
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
		Use:     "create [DATETIME]",
		Aliases: []string{"c", "new"},
		Short:   "Create a new journal entry. If no date provided, current time will be used.",
		Args:    cobra.MaximumNArgs(1),
		RunE:    create,
	}

	deleteCmd = &cobra.Command{
		Use:     "delete DATETIME",
		Aliases: []string{"d"},
		Short:   "Delete a journal entry.",
		Args:    cobra.ExactArgs(1),
		RunE:    delete,
	}

	listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "new"},
		Short:   "List journal entries, with an optional starting datetime.",
		Args:    cobra.NoArgs,
		RunE:    list,
	}

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync local db with latest Charm Cloud db.",
		Args:  cobra.NoArgs,
		RunE:  sync,
	}

	resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Delete local db and pull down fresh copy from Charm Cloud.",
		Args:  cobra.NoArgs,
		RunE:  reset,
	}
)

func create(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	model := client.CreateModel()
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		return err
	}

	return client.SaveEntry(cmd.Context(), db, time.Now(), model.Data)
}

func delete(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	return db.Delete([]byte(args[0]))
}

func list(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	// TODO: Check if online.
	if err := db.Sync(); err != nil {
		return err
	}

	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				in := fmt.Sprintf("# %s\n%s\n", item.Key(), string(v))

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

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func sync(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	return db.Sync()
}

func reset(cmd *cobra.Command, args []string) error {
	db, err := openKV()
	if err != nil {
		return err
	}

	return db.Reset()
}

func openKV() (*kv.KV, error) {
	return kv.OpenWithDefaults(dbName)
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
		resetCmd,
		syncCmd,
		cmd.LinkCmd("etu"),
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
