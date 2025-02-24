package components

import (
	"github.com/alex-laycalvert/ghtui/utils"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type spinnerModel struct {
	id      string
	spinner spinner.Model
}

func NewSpinnerComponent() spinnerModel {
	return spinnerModel{
		id: "spinner_" + uuid.NewString(),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
		),
	}
}

func (m spinnerModel) ID() string {
	return m.id
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case utils.FocusMsg:
		return m, m.spinner.Tick
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	return m.spinner.View()
}
