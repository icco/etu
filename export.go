package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all of your data as JSON to stdout or a file (-o).",
	Args:  cobra.NoArgs,
	RunE:  exportData,
}

func init() {
	exportCmd.Flags().StringP("output", "o", "", "write the export to this file instead of stdout")
}

func exportData(cmd *cobra.Command, _ []string) error {
	path, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	if path == "" {
		_, err := cfg.Export(cmd.Context(), os.Stdout)
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600) //nolint:gosec // G304: user-chosen output path
	if err != nil {
		return err
	}
	n, err := cfg.Export(cmd.Context(), f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		if rerr := os.Remove(path); rerr != nil {
			return errors.Join(err, rerr)
		}
		return err
	}
	fmt.Fprintf(os.Stderr, "Exported %d items to %s. Media URLs expire in one hour.\n", n, path)
	return nil
}
