package main

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/icco/etu/client"
	"github.com/spf13/cobra"
)

type searchModel struct {
	textInput   textinput.Model
	list        list.Model
	spinner     spinner.Model
	filtered    []*client.Post
	selected    *client.Post
	quitting    bool
	query       string
	showResults bool
	loading     bool
	loadErr     error
	cfg         *client.Config
}

type searchCompleteMsg struct {
	posts []*client.Post
	err   error
}

func performSearch(cfg *client.Config, query string) tea.Cmd {
	return func() tea.Msg {
		posts, err := cfg.SearchPosts(context.Background(), query, 50) // Limit to 50 results
		return searchCompleteMsg{posts: posts, err: err}
	}
}

func newSearchModel(cfg *client.Config) searchModel {
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

	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	// Create empty list initially - will be populated when user searches
	var items []list.Item
	buffer := 6
	maxSize := 10
	height := math.Min(float64(maxSize+buffer), float64(buffer))

	l := list.New(items, itemDelegate{}, 0, int(height))
	l.Title = "Search Results"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.Title = l.Styles.Title.Foreground(lipgloss.Color("170")).Bold(true)

	return searchModel{
		textInput:   ti,
		list:        l,
		spinner:     sp,
		showResults: false,
		loading:     false,
		cfg:         cfg,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case searchCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err
			return m, nil
		}
		m.filtered = msg.posts

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.textInput.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if !m.showResults {
				// User pressed enter on search query - perform search and show results
				query := strings.TrimSpace(m.textInput.Value())
				m.query = query
				m.loading = true
				m.loadErr = nil
				m.showResults = true

				// Start async search
				return m, tea.Batch(
					m.spinner.Tick,
					performSearch(m.cfg, query),
				)
			} else {
				// User pressed enter on a list item - select it
				if m.list.SelectedItem() != nil {
					item := m.list.SelectedItem().(listItem)
					m.selected = item.post
					m.quitting = true
					return m, tea.Quit
				}
			}
		}

		if !m.showResults {
			// Still in search input phase
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// In results list phase - update list when we have results
			if !m.loading && len(m.filtered) > 0 {
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
				m.textInput.Blur()
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m searchModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	if !m.showResults {
		// Show search prompt
		s.WriteString("\n  ")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("Search journal entries:"))
		s.WriteString("\n\n  ")
		s.WriteString(m.textInput.View())
		s.WriteString("\n")
	} else {
		// Show results list
		s.WriteString("\n  ")
		if m.loading {
			loadingText := fmt.Sprintf("%s Searching...", m.spinner.View())
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(loadingText))
		} else if m.loadErr != nil {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error searching: " + m.loadErr.Error()))
		} else {
			if m.query == "" {
				s.WriteString("All journal entries:")
			} else {
				s.WriteString(fmt.Sprintf("Search results for %q:", m.query))
			}
		}
		s.WriteString("\n\n")
		if !m.loading && len(m.filtered) > 0 {
			s.WriteString(m.list.View())
		} else if !m.loading && len(m.filtered) == 0 {
			s.WriteString("  No results found.\n")
		}
		s.WriteString("\n")
	}

	return docStyle.Render(s.String())
}

func searchPosts(cmd *cobra.Command, args []string) error {
	model := newSearchModel(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// If a post was selected, print it
	if finalModel.(searchModel).selected != nil {
		fmt.Println(finalModel.(searchModel).selected.Text)
	}

	return nil
}
