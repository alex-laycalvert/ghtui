package components

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
)

var (
	issuesTableStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62"))
	listItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF"))
	selectedListItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("#FFF"))
)

type IssuesListModel struct {
	width              int
	height             int
	issues             []*github.Issue
	viewportStartIndex int
	cursorIndex        int
}

type IssuesListUpdateWidthMsg struct {
	Width int
}

type IssuesListUpdateIssuesMsg struct {
	Issues []*github.Issue
}

type IssuesListResetViewportMsg struct{}

func NewIssuesListComponent(width int, height int) IssuesListModel {
	// Accounting for border width
	return IssuesListModel{width: width - 2, height: height - 2}
}

func (m IssuesListModel) GetSelectedIssue() *github.Issue {
	return m.issues[m.cursorIndex]
}

func (m IssuesListModel) Init() tea.Cmd {
	return nil
}

func (m IssuesListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "j":
			m.cursorIndex = min(len(m.issues)-1, m.cursorIndex+1)
			if m.cursorIndex >= m.viewportStartIndex+m.height {
				m.viewportStartIndex = m.viewportStartIndex + 1
				if m.viewportStartIndex+m.height > len(m.issues) {
					m.viewportStartIndex = len(m.issues) - m.height
				}
			}
			return m, nil
		case "k":
			m.cursorIndex = max(0, m.cursorIndex-1)
			if m.cursorIndex < m.viewportStartIndex {
				m.viewportStartIndex = m.cursorIndex
			}
			return m, nil
		case "H":
			m.cursorIndex = m.viewportStartIndex
			return m, nil
		case "L":
			m.cursorIndex = m.viewportStartIndex + m.height - 1
			return m, nil
		case "g":
			m.cursorIndex = 0
			m.viewportStartIndex = 0
			return m, nil
		case "G":
			m.cursorIndex = len(m.issues) - 1
			m.viewportStartIndex = max(0, len(m.issues)-m.height)
			return m, nil
		}
	case IssuesListUpdateWidthMsg:
		m.width = msg.Width
		return m, nil
	case IssuesListUpdateIssuesMsg:
		m.issues = msg.Issues
		return m, nil
	case IssuesListResetViewportMsg:
		m.viewportStartIndex = 0
		m.cursorIndex = 0
		return m, nil
	}

	return m, nil
}

func (m IssuesListModel) View() string {
	doc := strings.Builder{}

	for i := 0; i < m.height; i++ {
		if i+m.viewportStartIndex >= len(m.issues) {
			// Fill in blank space
			for j := i; j < m.height; j++ {
				doc.WriteString("\n")
			}
			break
		}

		issue := m.issues[i+m.viewportStartIndex]
		var listStyle lipgloss.Style
		if i+m.viewportStartIndex == m.cursorIndex {
			listStyle = selectedListItemStyle
		} else {
			listStyle = listItemStyle
		}
		listStyle = listStyle.Width(m.width)

		issueString := strconv.Itoa(*issue.Number) + " " + *issue.Title
		if len(issueString) >= m.width {
			issueString = issueString[:m.width-1]
		}
		doc.WriteString(listStyle.Render(issueString))
	}

	style := issuesTableStyle.
		Width(m.width).
		Height(m.height)
	return style.Render(doc.String())
}
