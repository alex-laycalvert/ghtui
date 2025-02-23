package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
)

type IssueModel struct {
	width  int
	height int

	issue    *github.Issue
	viewport viewport.Model
	renderer *glamour.TermRenderer
}

type IssueSetIssueMsg struct {
	Issue *github.Issue
}

func NewIssueComponent(issue *github.Issue, width int, height int) IssueModel {
	viewport := viewport.New(width, height)
	viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("62")).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		PaddingRight(2)
	renderWidth := width - viewport.Style.GetHorizontalFrameSize()
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(renderWidth),
	)
	return IssueModel{width, height, issue, viewport, renderer}
}

func (m IssueModel) Name() string {
	return "issue"
}

func (m IssueModel) GetIssue() *github.Issue {
	return m.issue
}

func (m IssueModel) Init() tea.Cmd {
	return nil
}

func (m IssueModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "g":
			if m.issue != nil {
				m.viewport.GotoTop()
			}
			return m, nil
		case "G":
			if m.issue != nil {
				m.viewport.GotoBottom()
			}
			return m, nil
		default:
			if m.issue != nil {
				m.viewport, _ = m.viewport.Update(msg)
			}
			return m, nil
		}
	case IssueSetIssueMsg:
		m.issue = msg.Issue
		if m.issue == nil {
			return m, nil
		}
		str, err := m.renderer.Render(*m.issue.Body)
		// TODO: error handling
		if err != nil {
			return m, nil
		}
		m.viewport.SetContent(str)
		return m, nil
	}

	return m, nil
}

func (m IssueModel) View() string {
	if m.issue == nil {
		return ""
	}

	return m.viewport.View()
}
