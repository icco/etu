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

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// styles holds the list's palette. Hex, not 256-color indexes, which the
// terminal's theme remaps; the pairs flip with the terminal background.
type styles struct {
	marker, date, tag, text, selText lipgloss.Style
	spinner, err                     lipgloss.Style
	key, desc, sep, dot, dotOff      lipgloss.Style
}

func newStyles(isDark bool) styles {
	ld := lipgloss.LightDark(isDark)
	fg := func(light, dark string) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(ld(lipgloss.Color(light), lipgloss.Color(dark)))
	}
	accent := fg("#B4632B", "#F5A97F")
	subtle := fg("#8A93A3", "#6E7787")

	return styles{
		marker:  accent.Bold(true),
		date:    fg("#55606E", "#A7B0C0"),
		tag:     fg("#1F7A6E", "#8BD5CA"),
		text:    fg("#1B1F26", "#F2F4F8"),
		selText: fg("#000000", "#FFFFFF").Bold(true),
		spinner: accent,
		err:     fg("#B3123A", "#ED8796"),
		key:     accent,
		desc:    fg("#4C5563", "#A7B0C0"),
		sep:     subtle,
		dot:     accent,
		dotOff:  subtle,
	}
}

type listItem struct {
	post *client.Post
}

func (i listItem) Title() string       { return i.post.CreatedAt.Format("2006-01-02 15:04") }
func (i listItem) Description() string { return i.post.Text }
func (i listItem) FilterValue() string { return i.post.Text }

type itemDelegate struct{ st styles }

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
	if !ok {
		return
	}

	marker, body := "  ", d.st.text
	if index == m.Index() {
		marker, body = d.st.marker.Render("\u276f "), d.st.selText
	}

	line := marker + d.st.date.Render(i.Title())
	if len(i.post.Tags) > 0 {
		line += " " + d.st.tag.Render("["+strings.Join(i.post.Tags, ", ")+"]")
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
	st       styles
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
	sp := spinner.New()
	sp.Spinner = spinner.Dot

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
	l.Styles.PaginationStyle = l.Styles.PaginationStyle.PaddingLeft(4)
	l.Styles.Title = l.Styles.Title.Bold(true)

	m := postListModel{
		list:    l,
		spinner: sp,
		loading: startLoading,
		cfg:     cfg,
		count:   count,
		title:   title,
	}
	// Assume dark until the terminal answers with its background color.
	m.applyStyles(newStyles(true))
	return m
}

// applyStyles paints the list with st. bubbles' own defaults for the help line
// and the pagination dots are near-black, and the paginator copies its dots at
// construction, so both are set here rather than through list.Styles.
func (m *postListModel) applyStyles(st styles) {
	m.st = st
	m.spinner.Style = st.spinner
	m.list.SetDelegate(itemDelegate{st: st})
	m.list.Paginator.ActiveDot = st.dot.Render("\u2022")
	m.list.Paginator.InactiveDot = st.dotOff.Render("\u2022")
	m.list.Help.Styles.ShortKey = st.key
	m.list.Help.Styles.FullKey = st.key
	m.list.Help.Styles.ShortDesc = st.desc
	m.list.Help.Styles.FullDesc = st.desc
	m.list.Help.Styles.ShortSeparator = st.sep
	m.list.Help.Styles.FullSeparator = st.sep
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

	case tea.BackgroundColorMsg:
		m.applyStyles(newStyles(msg.IsDark()))
		return m, nil

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
		s.WriteString(m.st.spinner.Render(loadingText))
		s.WriteString("\n")
	case m.loadErr != nil:
		s.WriteString("\n  ")
		s.WriteString(m.st.err.Render("Error: " + m.loadErr.Error()))
		s.WriteString("\n")
	case len(m.posts) > 0:
		s.WriteString(m.list.View())
	default:
		s.WriteString("\n  No entries found.\n")
	}

	v.SetContent(docStyle.Render(s.String()))
	return v
}
