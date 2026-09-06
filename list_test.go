package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/icco/etu/client"
)

func TestPostListModel(t *testing.T) {
	m := newPostListModel(nil, 5, "Entries", false)
	var mod tea.Model = m
	mod, _ = mod.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	mod, _ = mod.Update(postsLoadedMsg{posts: []*client.Post{{Text: "hello world"}}})
	v := mod.View()
	if !strings.Contains(v.Content, "hello world") {
		t.Fatalf("list did not render item: %q", v.Content)
	}
	if !v.AltScreen {
		t.Fatal("altscreen not set")
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
}
