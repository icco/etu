package main

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	gql "github.com/icco/graphql"
	"github.com/urfave/cli/v2"
)

type timerModel struct {
	stopwatch stopwatch.Model
	keymap    keymap
	help      help.Model
	quitting  bool
	cfg       *Config
	project   string
	start     time.Time
	sector    gql.WorkSector
	desc      string
}

func (m timerModel) Init() tea.Cmd {
	return m.stopwatch.Init()
}

func (m timerModel) View() string {
	// Note: you could further customize the time output by getting the
	// duration from m.stopwatch.Elapsed(), which returns a time.Duration, and
	// skip m.stopwatch.View() altogether.
	s := m.stopwatch.View() + "\n"
	if !m.quitting {
		s = "Elapsed: " + s
		s += m.helpView()
	}
	return s
}

func (m timerModel) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m timerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			m.cfg.Upload(context.Background(), m.start, time.Now(), m.sector, m.project, m.desc)
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.cfg.Upload(context.Background(), m.start, time.Now(), m.sector, m.project, m.desc)
			m.start = time.Now()
			return m, m.stopwatch.Reset()
		case key.Matches(msg, m.keymap.start):
			m.start = time.Now()
			m.keymap.stop.SetEnabled(!m.stopwatch.Running())
			m.keymap.start.SetEnabled(m.stopwatch.Running())
			return m, m.stopwatch.Toggle()
		case key.Matches(msg, m.keymap.stop):
			m.keymap.stop.SetEnabled(!m.stopwatch.Running())
			m.keymap.start.SetEnabled(m.stopwatch.Running())
			m.cfg.Upload(context.Background(), m.start, time.Now(), m.sector, m.project, m.desc)
			return m, m.stopwatch.Toggle()
		}
	}
	var cmd tea.Cmd
	m.stopwatch, cmd = m.stopwatch.Update(msg)
	return m, cmd
}

// Timer starts counting up.
func (cfg *Config) Timer(c *cli.Context) error {
	m := timerModel{
		cfg:       cfg,
		start:     time.Now(),
		sector:    gql.WorkSectorResearch,
		stopwatch: stopwatch.NewWithInterval(time.Millisecond),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.NewModel(),
	}

	m.keymap.start.SetEnabled(false)

	if err := tea.NewProgram(m).Start(); err != nil {
		return fmt.Errorf("bubbletea error: %w", err)
	}

	return nil
}
