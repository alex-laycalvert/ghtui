package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
	"golang.org/x/term"

	"github.com/alex-laycalvert/gtui/internal/components"
)

const START_ROW = 2

var selectedIssueGlamourStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	PaddingRight(2)

var issuesTableStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#000")).
	Foreground(lipgloss.Color("#FFF")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62"))

const borderWidth = 1

type model struct {
	repo        string
	issues      []*github.Issue
	currentPage int
	lastPage    int
	client      *github.Client

	width, height int

	selectedIssue         *github.Issue
	selectedIssueViewport viewport.Model
	selectedIssueRenderer *glamour.TermRenderer

	issuesComponent components.IssuesModel
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gtui <repo> <token>")
		os.Exit(1)
	}
	repo := os.Args[1]
	token := os.Args[2]

	client := github.NewClient(nil).WithAuthToken(token)

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	checkErr(err)

	m := model{
		repo:        repo,
		client:      client,
		width:       width,
		height:      height,
		currentPage: 1,

		// Height offset comes from header component
		issuesComponent: components.NewIssuesModel(width, height-1),
	}
	err = m.fetchIssues()
	checkErr(err)

	m.selectedIssueViewport = viewport.New(m.width/2-borderWidth*2, m.height-START_ROW-1)
	m.selectedIssueViewport.Style = selectedIssueGlamourStyle
	glamourRenderWidth := m.width/2 - m.selectedIssueViewport.Style.GetHorizontalFrameSize() - borderWidth*2
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	m.selectedIssueRenderer = renderer

	checkErr(err)

	_, err = tea.NewProgram(m).Run()
	checkErr(err)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "k":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "h":
			if m.selectedIssue != nil {
				return m, nil
			}

			m.issuesComponent.ResetViewport()
			m.selectedIssue = nil
			if m.currentPage == 1 {
				return m, nil
			}
			m.currentPage--
			err := m.fetchIssues()
			checkErr(err)
			return m, nil
		case "l":
			if m.selectedIssue != nil {
				return m, nil
			}

			m.issuesComponent.ResetViewport()
			m.selectedIssue = nil
			if m.currentPage == m.lastPage {
				return m, nil
			}
			m.currentPage++
			err := m.fetchIssues()
			checkErr(err)
			return m, nil
		case "g":
			if m.selectedIssue != nil {
				m.selectedIssueViewport.GotoTop()
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "G":
			if m.selectedIssue != nil {
				m.selectedIssueViewport.GotoBottom()
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "H":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "L":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.issuesComponent, _ = m.issuesComponent.Update(msg)
			return m, nil
		case "enter":
			m.selectedIssue = m.issuesComponent.GetSelectedIssue()
			m.issuesComponent.SetWidth(m.width / 2)
			str, err := m.selectedIssueRenderer.Render(*m.selectedIssue.Body)
			checkErr(err)
			m.selectedIssueViewport.SetContent(str)
			return m, nil
		case "esc":
			m.selectedIssue = nil
			m.issuesComponent.SetWidth(m.width)
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		m.Header(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.issuesComponent.View(),
			m.SelectedIssue(),
		))

	return view
}

func (m model) Header() string {
	return m.repo
}

func (m model) SelectedIssue() string {
	if m.selectedIssue != nil {
		return m.selectedIssueViewport.View()
	}
	return ""
}

func (m *model) fetchIssues() error {
	searchString := fmt.Sprintf("repo:%s is:open is:issue", m.repo)
	result, response, err := m.client.Search.Issues(context.Background(), searchString, &github.SearchOptions{
		Sort:        "created",
		Order:       "desc",
		ListOptions: github.ListOptions{Page: m.currentPage, PerPage: 50},
	})
	if err != nil {
		return err
	}
	m.issues = result.Issues
	m.issuesComponent.SetIssues(m.issues)
	m.lastPage = response.LastPage
	return nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
