package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single path", "/tmp/test.jpg", 1},
		{"multiple paths", "/tmp/a.jpg\n/tmp/b.png", 2},
		{"blank lines", "/tmp/a.jpg\n\n/tmp/b.png\n", 2},
		{"quoted paths", `"/tmp/my file.jpg"`, 1},
		{"single quoted", `'/tmp/my file.jpg'`, 1},
		{"whitespace only", "   \n  \n  ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePaths(tt.input)
			if len(got) != tt.want {
				t.Errorf("parsePaths(%q) returned %d paths, want %d", tt.input, len(got), tt.want)
			}
			for _, p := range got {
				if !filepath.IsAbs(p) {
					t.Errorf("parsePaths(%q) returned non-absolute path: %s", tt.input, p)
				}
			}
		})
	}
}

func TestParsePathsStripsQuotes(t *testing.T) {
	got := parsePaths(`"/tmp/test.jpg"`)
	if len(got) != 1 {
		t.Fatalf("expected 1 path, got %d", len(got))
	}
	if got[0] != "/tmp/test.jpg" {
		t.Errorf("expected /tmp/test.jpg, got %s", got[0])
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"long", "hello world", 8, "hello..."},
		{"newlines replaced", "hello\nworld", 20, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"zero", 0, "0.0h"},
		{"negative", -5 * time.Minute, "0.0h"},
		{"minutes", 30 * time.Minute, "0.5h"},
		{"hours", 3 * time.Hour, "3.0h"},
		{"day boundary", 24 * time.Hour, "24.0h"},
		{"over a day", 25 * time.Hour, "1.0d"},
		{"days", 72 * time.Hour, "3.0d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.dur)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}
