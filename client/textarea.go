package client

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
}

type errMsg error

type model struct {
	textarea textarea.Model
	err      error
	Data     []byte
}

// CreateModel generates a text area for use in bubbletea.
func CreateModel() model {
	ti := textarea.New()
	ti.Placeholder = "Once upon a time..."
	ti.Focus()

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
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

func (m model) View() string {
	return fmt.Sprintf(
		"What are you working on?\n\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+c to quit)",
	) + "\n\n"
}
