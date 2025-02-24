package components

import (
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

	content  string
	viewport viewport.Model
	renderer *glamour.TermRenderer
}

type MarkdownViewerSetContentMsg struct {
	Content string
}

func NewMarkdownViewerComponent(width int, height int, style lipgloss.Style) markdownViewerModel {
	viewport := viewport.New(width, height)
	viewport.Style = style
	renderWidth := width - viewport.Style.GetHorizontalFrameSize()
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(renderWidth),
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

func (m markdownViewerModel) View() string {
	return m.viewport.View()
}

func (m *markdownViewerModel) setContent(content string) {
	str, err := m.renderer.Render(content)
	// TODO: error handling
	if err != nil {
		return
	}
	m.viewport.SetContent(str)
}
