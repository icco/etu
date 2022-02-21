package main

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	gql "github.com/icco/graphql"
	"github.com/urfave/cli/v2"
)

const timeout = time.Second * 5

type pomoModel struct {
	timer    timer.Model
	keymap   keymap
	help     help.Model
	quitting bool
	cfg      *Config
	project  string
	start    time.Time
	sector   gql.Sector
	desc     string
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func (m pomoModel) Init() tea.Cmd {
	return m.timer.Init()
}

func (m pomoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		m.cfg.Upload(context.Background(), m.start, time.Now(), m.sectory, m.project, m.desc)
		return m, tea.Quit

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			m.cfg.Upload(context.Background(), m.start, time.Now(), m.sectory, m.project, m.desc)
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.timer.Timeout = timeout
		case key.Matches(msg, m.keymap.start, m.keymap.stop):
			return m, m.timer.Toggle()
		}
	}

	return m, nil
}

func (m pomoModel) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m pomoModel) View() string {
	s := m.timer.View()

	if m.timer.Timedout() {
		s = "All done!"
	}
	s += "\n"
	if !m.quitting {
		s = "Exiting in " + s
		s += m.helpView()
	}
	return s
}

// Pomodoro creates a 25 minute countdown to work to.
func (cfg *Config) Pomodoro(c *cli.Context) error {
	timeout := time.Minute * 25

	m := pomoModel{
		cfg:   cfg,
		start: time.Now(),
		timer: timer.NewWithInterval(timeout, time.Millisecond),
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
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.NewModel(),
	}
	m.keymap.start.SetEnabled(false)

	if err := tea.NewProgram(m).Start(); err != nil {
		return err
	}

	return nil
}
