package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/icco/etu/client"
)

func TestPostListModel(t *testing.T) {
	m := newPostListModel(nil, 5, "Entries", false)
	var mod tea.Model = m
	mod, _ = mod.Update(tea.BackgroundColorMsg{Color: lipgloss.Color("#ffffff")})
	if got := mod.(postListModel).st.text.GetForeground(); got != lipgloss.Color("#1B1F26") {
		t.Fatalf("light terminal did not flip the palette: got %v", got)
	}
	mod, _ = mod.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	mod, _ = mod.Update(postsLoadedMsg{posts: []*client.Post{
		{Text: "hello world"},
		{Text: "first line\n\nsecond line", Tags: []string{"work"}},
	}})
	v := mod.View()
	if !strings.Contains(v.Content, "hello world") {
		t.Fatalf("list did not render item: %q", v.Content)
	}
	if !v.AltScreen {
		t.Fatal("altscreen not set")
	}
	if !strings.Contains(v.Content, "first line second line") {
		t.Fatalf("multi-line entry not collapsed to one row: %q", v.Content)
	}

	mod, cmd := mod.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter produced no cmd (expected tea.Quit)")
	}
	if mod.(postListModel).selected == nil {
		t.Fatal("enter did not select an item")
	}

	if _, cmd := mod.Update(tea.KeyPressMsg{Code: 'q', Text: "q"}); cmd == nil {
		t.Fatal("q produced no cmd (expected tea.Quit)")
	}

	if _, cmd := mod.Update(tea.KeyPressMsg{Code: 'v', Text: "v"}); cmd != nil {
		t.Fatal("v quit the list; bubbles' default quit binding leaked through")
	}
}
