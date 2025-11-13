package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/icco/etu/client"
	"github.com/icco/etu/search"
	"github.com/spf13/cobra"
)

type searchModel struct {
	textInput textinput.Model
	list      list.Model
	allPosts  []*client.Post
	filtered  []*client.Post
	selected  *client.Post
	quitting  bool
}

func newSearchModel(posts []*client.Post) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search journal entries..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	// Create initial list with all posts
	var items []list.Item
	for _, p := range posts {
		items = append(items, listItem{post: p})
	}

	buffer := 6
	maxSize := 10
	height := math.Min(float64(maxSize+buffer), float64(len(items)+buffer))

	l := list.New(items, itemDelegate{}, 0, int(height))
	l.Title = "Search Results"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)

	return searchModel{
		textInput: ti,
		list:      l,
		allPosts:  posts,
		filtered:  posts,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// If we have a selected item, print it and exit
			if m.list.SelectedItem() != nil {
				item := m.list.SelectedItem().(listItem)
				m.selected = item.post
				m.quitting = true
				return m, tea.Quit
			}
		}

		// Update text input
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		// Perform fuzzy search on text input change
		query := m.textInput.Value()
		m.filtered = search.SearchPosts(query, m.allPosts)

		// Update list items
		var items []list.Item
		for _, p := range m.filtered {
			items = append(items, listItem{post: p})
		}

		buffer := 6
		maxSize := 10
		height := math.Min(float64(maxSize+buffer), float64(len(items)+buffer))
		m.list = list.New(items, itemDelegate{}, 0, int(height))
		m.list.Title = fmt.Sprintf("Search Results (%d)", len(m.filtered))
		m.list.SetShowStatusBar(false)
		m.list.SetFilteringEnabled(false)
		m.list.SetShowTitle(true)
		m.list.SetShowHelp(true)
		m.list.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m searchModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(docStyle.Render(m.list.View()))
	return b.String()
}

func searchPosts(cmd *cobra.Command, args []string) error {
	// Fetch a large number of posts for searching
	// Notion API has a limit, so we'll fetch 100 posts
	entries, err := cfg.ListPosts(cmd.Context(), 100)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No journal entries found.")
		return nil
	}

	model := newSearchModel(entries)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	// If a post was selected, print it
	if model.selected != nil {
		fmt.Println(model.selected.Text)
	}

	return nil
}
