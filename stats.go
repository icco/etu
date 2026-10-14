package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.icco.me/etu/client"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show journal stats (blips, tags, words written).",
	Args:  cobra.NoArgs,
	RunE:  showStats,
}

func showStats(cmd *cobra.Command, _ []string) error {
	stats, err := cfg.GetStats(cmd.Context())
	if err != nil {
		return err
	}
	fmt.Print(formatStats(stats))
	return nil
}

// formatStats renders stats as one metric per line.
func formatStats(s client.Stats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Blips: %d\n", s.TotalBlips)
	fmt.Fprintf(&b, "Tags: %d\n", s.UniqueTags)
	fmt.Fprintf(&b, "Words written: %d\n", s.WordsWritten)
	return b.String()
}
