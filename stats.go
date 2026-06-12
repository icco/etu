package main

import (
	"fmt"
	"strings"

	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show journal stats (blips, tags, words written).",
	Args:  cobra.NoArgs,
	RunE:  showStats,
}

func showStats(cmd *cobra.Command, _ []string) error {
	personal, err := cfg.GetStats(cmd.Context(), false)
	if err != nil {
		return err
	}

	global, err := cmd.Flags().GetBool("global")
	if err != nil {
		return err
	}

	var community *client.Stats
	if global {
		stats, err := cfg.GetStats(cmd.Context(), true)
		if err != nil {
			return err
		}
		community = &stats
	}

	fmt.Print(formatStats(personal, community))
	return nil
}

// formatStats renders stats as one metric per line; a non-nil community adds
// a second "Community" block.
func formatStats(personal client.Stats, community *client.Stats) string {
	var b strings.Builder
	writeBlock := func(s client.Stats) {
		fmt.Fprintf(&b, "Blips: %d\n", s.TotalBlips)
		fmt.Fprintf(&b, "Tags: %d\n", s.UniqueTags)
		fmt.Fprintf(&b, "Words written: %d\n", s.WordsWritten)
	}
	writeBlock(personal)
	if community != nil {
		b.WriteString("\nCommunity\n")
		writeBlock(*community)
	}
	return b.String()
}
