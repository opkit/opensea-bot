package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type Item struct {
	Placeholder string
}
type model struct {
	Input     []Item
	err       error
	textInput textinput.Model
	index     int
	Value     map[int]string
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	return model{
		Input: []Item{
			{Placeholder: "Token Address"},
			{Placeholder: "Token Network"},
		},
		err:       nil,
		Value:     map[int]string{},
		index:     0,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	m.textInput.Placeholder = m.Input[m.index].Placeholder
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.index < len(m.Input) {
				m.Value[m.index] = strings.TrimSpace(m.textInput.Value())
			}
			m.index++
			if m.index == len(m.Input) {
				return m, tea.Quit
			}
			m.textInput.Reset()
			return m, cmd
		}

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var value []string
	for idx, v := range m.Value {
		value = append(value, fmt.Sprintf("%s: %s", m.Input[idx].Placeholder, v))
	}
	if m.index != len(m.Input) {
		value = append(value, fmt.Sprintf("%s: %s", m.Input[m.index].Placeholder, m.textInput.View()))
	}
	str := strings.Join(value, "\n")
	return fmt.Sprintf("%s\n\n%s", str, "(esc to quit)") + "\n"
}
