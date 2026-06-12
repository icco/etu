package main

import (
	"testing"

	"github.com/icco/etu/client"
)

func TestFormatStats(t *testing.T) {
	personal := client.Stats{TotalBlips: 10, UniqueTags: 5, WordsWritten: 1234}
	community := client.Stats{TotalBlips: 1000, UniqueTags: 200, WordsWritten: 99999}

	tests := []struct {
		name      string
		personal  client.Stats
		community *client.Stats
		want      string
	}{
		{
			"personal only",
			personal,
			nil,
			"Blips: 10\nTags: 5\nWords written: 1234\n",
		},
		{
			"with community block",
			personal,
			&community,
			"Blips: 10\nTags: 5\nWords written: 1234\n\nCommunity\nBlips: 1000\nTags: 200\nWords written: 99999\n",
		},
		{
			"zero values",
			client.Stats{},
			nil,
			"Blips: 0\nTags: 0\nWords written: 0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStats(tt.personal, tt.community)
			if got != tt.want {
				t.Errorf("formatStats() = %q, want %q", got, tt.want)
			}
		})
	}
}
