package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/charm/client"
	"github.com/charmbracelet/charm/cmd"
	"github.com/charmbracelet/charm/kv"
	"github.com/charmbracelet/charm/ui/common"
	"github.com/dgraph-io/badger/v3"
	"github.com/muesli/roff"
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

	newCmd = &cobra.Command{
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
		Args:    cobra.NoArgs(),
		RunE:    list,
	}

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync local db with latest Charm Cloud db.",
		Args:  cobra.NoArgs(),
		RunE:  sync,
	}

	resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Delete local db and pull down fresh copy from Charm Cloud.",
		Args:  cobra.NoArgs(),
		RunE:  reset,
	}
)

func create(cmd *cobra.Command, args []string) error {
	k, n, err := keyParser(args[0])
	if err != nil {
		return err
	}
	db, err := openKV(n)
	if err != nil {
		return err
	}
	if len(args) == 2 {
		return db.Set(k, []byte(args[1]))
	}
	return db.SetReader(k, os.Stdin)
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
		if keysIterate {
			opts.PrefetchValues = false
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if keysIterate {
				printFromKV(pf, k)
				continue
			}
			err := item.Value(func(v []byte) error {
				if valuesIterate {
					printFromKV(pf, v)
				} else {
					printFromKV(pf, k, v)
				}
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
	db, err := openKV(n)
	if err != nil {
		return err
	}
	return db.Sync()
}

func reset(cmd *cobra.Command, args []string) error {
	n, err := nameFromArgs(args)
	if err != nil {
		return err
	}
	db, err := openKV(n)
	if err != nil {
		return err
	}
	return db.Reset()
}

func nameFromArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	_, n, err := keyParser(args[0])
	if err != nil {
		return "", err
	}
	return n, nil
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
