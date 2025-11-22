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

type postListModel struct {
	list     list.Model
	spinner  spinner.Model
	loading  bool
	loadErr  error
	posts    []*client.Post
	selected *client.Post
	cfg      *client.Config
	count    int
	title    string
	query    string
	quitting bool
}

type postsLoadedMsg struct {
	posts []*client.Post
	err   error
}

func loadPosts(cfg *client.Config, count int, query string) tea.Cmd {
	return func() tea.Msg {
		var posts []*client.Post
		var err error
		if query != "" {
			posts, err = cfg.SearchPosts(context.Background(), query, count)
		} else {
			posts, err = cfg.ListPosts(context.Background(), count)
		}
		return postsLoadedMsg{posts: posts, err: err}
	}
}

func newPostListModel(cfg *client.Config, count int, title string, startLoading bool) postListModel {
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
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.Title = l.Styles.Title.Foreground(lipgloss.Color("170")).Bold(true)

	return postListModel{
		list:    l,
		spinner: sp,
		loading: startLoading,
		cfg:     cfg,
		count:   count,
		title:   title,
	}
}

func (m postListModel) Init() tea.Cmd {
	if !m.loading {
		return nil
	}
	// Start loading posts asynchronously
	return tea.Batch(
		m.spinner.Tick,
		loadPosts(m.cfg, m.count, m.query),
	)
}

func (m postListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case postsLoadedMsg:
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
			if m.query != "" {
				m.list.Title = fmt.Sprintf("Search Results (%d)", len(m.posts))
			} else {
				m.list.Title = m.title
			}
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
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.list.SelectedItem() != nil {
				// User selected an item
				item := m.list.SelectedItem().(listItem)
				m.selected = item.post
				m.quitting = true
				return m, tea.Quit
			}
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

func (m postListModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	if m.loading {
		var loadingText string
		if m.query != "" {
			loadingText = fmt.Sprintf("%s Searching...", m.spinner.View())
		} else {
			loadingText = fmt.Sprintf("%s Loading journal entries...", m.spinner.View())
		}
		s.WriteString("\n  ")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(loadingText))
		s.WriteString("\n")
	} else if m.loadErr != nil {
		s.WriteString("\n  ")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: " + m.loadErr.Error()))
		s.WriteString("\n")
	} else if len(m.posts) > 0 {
		s.WriteString(m.list.View())
	} else {
		s.WriteString("\n  No entries found.\n")
	}

	return docStyle.Render(s.String())
}
