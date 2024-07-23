package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type model struct {
	textarea textarea.Model
	err      error
	Data     []byte
}

// createModel generates a text area for use in bubbletea.
func createModel() *model {
	ta := textarea.New()
	ta.Placeholder = "What are you working on? (ctrl+d to save, ctrl+c to quit)"
	ta.Focus()

	ta.SetWidth(120)
	ta.SetHeight(5)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	return &model{
		textarea: ta,
		err:      nil,
	}
}

func (m *model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlD:
			m.Data = []byte(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	return m.textarea.View()
}
