package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type MarkdownViewerModel struct {
	width  int
	height int

	content  string
	viewport viewport.Model
	renderer *glamour.TermRenderer
}

type MarkdownViewerSetContentMsg struct {
	Content string
}

func NewMarkdownViewerComponent(width int, height int, style lipgloss.Style) MarkdownViewerModel {
	viewport := viewport.New(width, height)
	viewport.Style = style
	renderWidth := width - viewport.Style.GetHorizontalFrameSize()
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(renderWidth),
	)
	m := MarkdownViewerModel{
		width:    width,
		height:   height,
		viewport: viewport,
		renderer: renderer,
	}
	return m
}

func (m MarkdownViewerModel) Init() tea.Cmd {
	return nil
}

func (m MarkdownViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.setContent(msg.Content)
		return m, nil
	}

	return m, nil
}

func (m MarkdownViewerModel) View() string {
	return m.viewport.View()
}

func (m *MarkdownViewerModel) setContent(content string) {
	str, err := m.renderer.Render(content)
	// TODO: error handling
	if err != nil {
		return
	}
	m.viewport.SetContent(str)
}
