package main

import (
	"testing"

	"github.com/icco/etu/client"
)

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name string
		tags []client.Tag
		want string
	}{
		{"empty", nil, ""},
		{"single", []client.Tag{{Name: "work", Count: 3}}, "work (3)\n"},
		{
			"sorted by count desc",
			[]client.Tag{
				{Name: "life", Count: 2},
				{Name: "work", Count: 5},
			},
			"work (5)\nlife (2)\n",
		},
		{
			"ties broken by name asc",
			[]client.Tag{
				{Name: "zebra", Count: 2},
				{Name: "apple", Count: 2},
				{Name: "work", Count: 5},
			},
			"work (5)\napple (2)\nzebra (2)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTags(tt.tags)
			if got != tt.want {
				t.Errorf("formatTags(%v) = %q, want %q", tt.tags, got, tt.want)
			}
		})
	}
}
