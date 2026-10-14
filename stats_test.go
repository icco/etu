package main

import (
	"testing"

	"go.icco.me/etu/client"
)

func TestFormatStats(t *testing.T) {
	tests := []struct {
		name  string
		stats client.Stats
		want  string
	}{
		{"values", client.Stats{TotalBlips: 10, UniqueTags: 5, WordsWritten: 1234}, "Blips: 10\nTags: 5\nWords written: 1234\n"},
		{"zero values", client.Stats{}, "Blips: 0\nTags: 0\nWords written: 0\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatStats(tt.stats); got != tt.want {
				t.Errorf("formatStats() = %q, want %q", got, tt.want)
			}
		})
	}
}
