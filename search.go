package main

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/icco/etu/client"
	"github.com/icco/etu/search"
	"github.com/spf13/cobra"
)

var (
	searchInputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		MarginBottom(1)
)

type searchModel struct {
	textInput  textinput.Model
	list       list.Model
	searchable *search.SearchablePosts
	filtered   []*client.Post
	selected   *client.Post
	quitting   bool
	width      int
	height     int
}

func newSearchModel(posts []*client.Post) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search journal entries..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	// Style the textinput components
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	// Pre-compute searchable posts for performance
	searchable := search.NewSearchablePosts(posts)

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
	l.Styles.Title = l.Styles.Title.Foreground(lipgloss.Color("170")).Bold(true)

	return searchModel{
		textInput:  ti,
		list:       l,
		searchable: searchable,
		filtered:   posts,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.textInput.Width = msg.Width - 4 // Account for margins
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
		oldQuery := m.textInput.Value()
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		// Only update list if query changed
		newQuery := m.textInput.Value()
		if oldQuery != newQuery {
			// Perform fuzzy search on text input change
			m.filtered = m.searchable.Search(newQuery)

			// Update list items efficiently using SetItems
			var items []list.Item
			for _, p := range m.filtered {
				items = append(items, listItem{post: p})
			}

			buffer := 6
			maxSize := 10
			height := math.Min(float64(maxSize+buffer), float64(len(items)+buffer))
			m.list.SetItems(items)
			m.list.SetHeight(int(height))
			m.list.Title = fmt.Sprintf("Search Results (%d)", len(m.filtered))
		}
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

	// Style the input with a border
	inputView := searchInputStyle.Render(m.textInput.View())

	// Combine input and list with proper spacing
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		"",
		m.list.View(),
	)

	return docStyle.Render(view)
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
