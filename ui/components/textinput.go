package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TextInputComponent struct {
	width int

	isFocused bool
	cursor    cursor.Model

	label string
	value string
}

type TextInputSubmitMsg struct {
	Value string
}

type TextInputClearMsg struct{}

func NewTextInputComponent(label string, width int) TextInputComponent {
	cursor := cursor.New()
	cursor.SetChar(" ")
	return TextInputComponent{label: label, width: width, cursor: cursor}
}

func (m TextInputComponent) Init() tea.Cmd {
	return nil
}

func (m TextInputComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			return m, func() tea.Msg {
				return TextInputSubmitMsg{Value: m.value}
			}
		case "ctrl+w":
			// Kill prev word
			index := 0
			for i := len(m.value) - 1; i >= 0; i-- {
				if m.value[i] == ' ' {
					index = i
					break
				}
			}
			if len(m.value) > 0 && index == len(m.value)-1 {
				index--
			}

			if index == 0 {
				m.value = ""
			} else {
				m.value = m.value[:index+1]
			}
			return m, nil
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
			return m, nil
		default:
			m.value += keypress
			return m, nil
		}
	case TextInputClearMsg:
		m.value = ""
		return m, nil
	case ComponentUpdateSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}
		return m, nil
	case tea.FocusMsg:
		m.isFocused = true
		return m, tea.Sequence(
			m.cursor.Focus(),
			m.cursor.BlinkCmd(),
		)
	case tea.BlurMsg:
		m.isFocused = false
		m.cursor.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.cursor, cmd = m.cursor.Update(msg)
		return m, cmd
	}
}

func (m TextInputComponent) View() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			fmt.Sprintf("%s: %s", m.label, m.value),
			m.cursor.View(),
		))
}
