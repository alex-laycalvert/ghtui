package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
	"golang.org/x/term"

	"github.com/alex-laycalvert/ghtui/internal/components"
	"github.com/alex-laycalvert/ghtui/internal/pages/issues"
	"github.com/alex-laycalvert/ghtui/internal/pages/repo"
)

type model struct {
	width  int
	height int

	client *github.Client
	repo   string

	pages components.ComponentGroup
}

const (
	repoPage   components.ComponentName = "Repo"
	issuesPage components.ComponentName = "Issues"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gtui <repo> <token>")
		os.Exit(1)
	}
	repoName := os.Args[1]
	token := os.Args[2]

	client := github.NewClient(nil).WithAuthToken(token)

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	checkErr(err)

	pageWidth := width - 6
	pageHeight := height - 6

	pages := components.NewComponentGroup(
		components.NameComponent(
			repoPage,
			repo.NewRepoPage(client, repoName, pageWidth, pageHeight),
		),
		components.NameComponent(
			issuesPage,
			issues.NewIssuesPage(client, repoName, pageWidth, pageHeight),
		),
	)

	m := model{
		client: client,
		repo:   repoName,
		width:  width,
		height: height,
		pages:  pages,
	}

	_, err = tea.NewProgram(m).Run()
	checkErr(err)
}

func (m model) Init() tea.Cmd {
	return m.pages.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.pages.FocusNext()
			return m, m.pages.GetFocusedComponent().Init()
		case "shift+tab":
			m.pages.FocusPrevious()
			return m, m.pages.GetFocusedComponent().Init()
		default:
			return m, m.pages.UpdateFocused(msg)
		}
	default:
		return m, m.pages.UpdateFocused(msg)
	}
}

func tabBorderStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(highlightColor).
		Padding(0, 1)
	return style
}

var (
	docStyle         = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle = tabBorderStyle()
	activeTabStyle   = inactiveTabStyle.
				Bold(true)
	windowStyle = lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Border(lipgloss.RoundedBorder())
)

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	pages := m.pages.GetComponents()
	currentPage := m.pages.GetFocusedComponent()
	for _, t := range pages {
		var style lipgloss.Style
		isActive := t.Name() == currentPage.Name()
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(string(t.Name())))
	}

	header := lipgloss.NewStyle().
		MarginLeft(1).
		Padding(1).
		Render(m.repo)
	row := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...),
		header,
	)
	doc.WriteString(row + "\n")
	doc.WriteString(
		windowStyle.
			Render(currentPage.View()),
	)
	return docStyle.
		Width(m.width).
		Height(m.height).
		Render(doc.String())
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
