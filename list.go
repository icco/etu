// Package main implements the etu CLI: a personal command-line journal.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/icco/etu/client"
)

const (
	listBuffer  = 6
	listMaxSize = 10
)

var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	markerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	dateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	tagStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
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

	marker, body := "  ", textStyle
	if index == m.Index() {
		marker, body = markerStyle.Render("\u276f "), selTextStyle
	}

	line := marker + dateStyle.Render(i.Title())
	if len(i.post.Tags) > 0 {
		line += " " + tagStyle.Render("["+strings.Join(i.post.Tags, ", ")+"]")
	}
	line += "  " + body.Render(oneLine(i.Description()))

	// The list allots each item a single row, so truncate to its width.
	if width := m.Width(); width > 0 {
		line = ansi.Truncate(line, width, "\u2026")
	}

	if _, err := fmt.Fprint(w, line); err != nil {
		log.Printf("list render: %v", err)
	}
}

// oneLine collapses an entry's whitespace so a multi-line entry stays on its row.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
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
	sp.Style = spinnerStyle

	// Create empty list initially - will be populated when data loads
	var items []list.Item
	l := list.New(items, itemDelegate{}, 0, listBuffer)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(true)
	// bubbles v2 binds the list's quit key to "v" and labels it "select".
	l.KeyMap.Quit = key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q", "quit"))
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))}
	}
	l.Styles.PaginationStyle = list.DefaultStyles(true).PaginationStyle.PaddingLeft(4)
	l.Styles.Title = l.Styles.Title.Bold(true)

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

			m.list.SetItems(items)
			m.list.SetHeight(int(math.Min(float64(listMaxSize+listBuffer), float64(len(items)+listBuffer))))
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
		frame, _ := docStyle.GetFrameSize()
		m.list.SetWidth(msg.Width - frame)
		return m, nil

	case tea.KeyPressMsg:
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

func (m postListModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	if m.quitting {
		return v
	}

	var s strings.Builder

	switch {
	case m.loading:
		var loadingText string
		if m.query != "" {
			loadingText = fmt.Sprintf("%s Searching...", m.spinner.View())
		} else {
			loadingText = fmt.Sprintf("%s Loading journal entries...", m.spinner.View())
		}
		s.WriteString("\n  ")
		s.WriteString(spinnerStyle.Render(loadingText))
		s.WriteString("\n")
	case m.loadErr != nil:
		s.WriteString("\n  ")
		s.WriteString(errStyle.Render("Error: " + m.loadErr.Error()))
		s.WriteString("\n")
	case len(m.posts) > 0:
		s.WriteString(m.list.View())
	default:
		s.WriteString("\n  No entries found.\n")
	}

	v.SetContent(docStyle.Render(s.String()))
	return v
}
