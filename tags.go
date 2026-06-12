package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"t"},
	Short:   "List all tags with usage counts.",
	Args:    cobra.NoArgs,
	RunE:    listTags,
}

func listTags(cmd *cobra.Command, _ []string) error {
	tags, err := cfg.ListTags(cmd.Context())
	if err != nil {
		return err
	}

	fmt.Print(formatTags(tags))
	return nil
}

// formatTags renders tags one per line as "name (count)", sorted by count
// descending then name ascending.
func formatTags(tags []client.Tag) string {
	sorted := make([]client.Tag, len(tags))
	copy(sorted, tags)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}
		return sorted[i].Name < sorted[j].Name
	})

	var b strings.Builder
	for _, t := range sorted {
		fmt.Fprintf(&b, "%s (%d)\n", t.Name, t.Count)
	}
	return b.String()
}
