package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/icco/etu/client"
)

var (
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type listItem struct {
	post *client.Post
}

func (i listItem) Title() string       { return i.post.CreatedAt.Format("2006-01-02 15:04") }
func (i listItem) Description() string { return i.post.Text }
func (i listItem) FilterValue() string { return i.post.Text }

type listModel struct {
	list    list.Model
	spinner spinner.Model
	loading bool
	loadErr error
	posts   []*client.Post
	cfg     *client.Config
	count   int
}

type listCompleteMsg struct {
	posts []*client.Post
	err   error
}

func performList(cfg *client.Config, count int) tea.Cmd {
	return func() tea.Msg {
		posts, err := cfg.ListPosts(context.Background(), count)
		return listCompleteMsg{posts: posts, err: err}
	}
}

func newListModel(cfg *client.Config, count int) listModel {
	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	// Create empty list initially - will be populated when data loads
	var items []list.Item
	buffer := 6
	maxSize := 10
	height := math.Min(float64(maxSize+buffer), float64(buffer))

	l := list.New(items, itemDelegate{}, 0, int(height))
	l.Title = "Interstitial Notes"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.Title = l.Styles.Title.Foreground(lipgloss.Color("170")).Bold(true)

	return listModel{
		list:    l,
		spinner: sp,
		loading: true,
		cfg:     cfg,
		count:   count,
	}
}

func (m listModel) Init() tea.Cmd {
	// Start loading posts asynchronously
	return tea.Batch(
		m.spinner.Tick,
		performList(m.cfg, m.count),
	)
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case listCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err
			return m, nil
		}
		m.posts = msg.posts

		// Update list with results
		if len(m.posts) > 0 {
			var items []list.Item
			for _, p := range m.posts {
				items = append(items, listItem{post: p})
			}

			buffer := 6
			maxSize := 10
			height := math.Min(float64(maxSize+buffer), float64(len(items)+buffer))
			m.list.SetItems(items)
			m.list.SetHeight(int(height))
			m.list.Title = "Interstitial Notes"
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			return m, tea.Quit
		}
	}

	// Only update list if we have loaded posts
	if !m.loading && len(m.posts) > 0 {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m listModel) View() string {
	var s strings.Builder

	if m.loading {
		loadingText := fmt.Sprintf("%s Loading journal entries...", m.spinner.View())
		s.WriteString("\n  ")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(loadingText))
		s.WriteString("\n")
	} else if m.loadErr != nil {
		s.WriteString("\n  ")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error loading entries: " + m.loadErr.Error()))
		s.WriteString("\n")
	} else if len(m.posts) > 0 {
		s.WriteString(m.list.View())
	} else {
		s.WriteString("\n  No entries found.\n")
	}

	return docStyle.Render(s.String())
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("> %s - %s", i.Title(), i.Description())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(s...)
		}
	}

	fmt.Fprint(w, fn(str))
}
