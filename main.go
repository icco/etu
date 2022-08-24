package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/charm/cmd"
	"github.com/charmbracelet/charm/kv"
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

	db.Set([]byte(time.Now().Format(time.RFC3339)), model.Data)

	return nil
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
		listCmd,
		syncCmd,
		resetCmd,
		cmd.LinkCmd("etu"),
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
