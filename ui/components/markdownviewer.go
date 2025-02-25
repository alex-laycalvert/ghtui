package components

import (
	"github.com/alex-laycalvert/ghtui/utils"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type markdownViewerModel struct {
	id     string
	width  int
	height int

	style    lipgloss.Style
	content  string
	viewport viewport.Model
	renderer *glamour.TermRenderer

	updateTimes int
}

type MarkdownViewerSetContentMsg struct {
	Content string
}

type markdownViewerUpdateMarkdownMsg struct {
	id       string
	renderer *glamour.TermRenderer
}

func NewMarkdownViewerComponent(width int, height int, style lipgloss.Style) markdownViewerModel {
	viewport := viewport.New(width, height)
	viewport.Style = style
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	m := markdownViewerModel{
		id:       "markdownViewer_" + uuid.NewString(),
		width:    width,
		height:   height,
		viewport: viewport,
		renderer: renderer,
	}
	return m
}

func (m markdownViewerModel) ID() string {
	return m.id
}

func (m markdownViewerModel) Init() tea.Cmd {
	return nil
}

func (m markdownViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case utils.FocusMsg:
		if m.id != msg.ID {
			return m, nil
		}
		return m, nil
	case utils.UpdateSizeMsg:
		if m.id != msg.ID {
			return m, nil
		}

		if msg.Width == 0 && msg.Height == 0 {
			return m, nil
		}

		if msg.Width > 0 {
			m.width = msg.Width
		}
		if msg.Height > 0 {
			m.height = msg.Height
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height

		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		default:
			m.viewport, _ = m.viewport.Update(msg)
			return m, nil
		}
	case MarkdownViewerSetContentMsg:
		str, _ := m.renderer.Render(msg.Content)
		m.viewport.SetContent(str)
		return m, nil
	}

	return m, nil
}

func (m markdownViewerModel) View() string {
	return m.viewport.View()
}
