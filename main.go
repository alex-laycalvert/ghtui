package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
	"golang.org/x/term"
)

const START_ROW = 3

type model struct {
	repo               string
	issues             []*github.Issue
	currentPage        int
	lastPage           int
	cursorIndex        int
	viewportStartIndex int
	selectedIssueIndex int
	client             *github.Client

	width, height int
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
		repo:               repo,
		client:             client,
		width:              width,
		height:             height,
		currentPage:        1,
		selectedIssueIndex: -1,
	}
	err = m.fetchIssues()
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
			m.cursorIndex = min(m.cursorIndex+1, len(m.issues)-1)
			if m.cursorIndex > m.viewportStartIndex+m.height-START_ROW-2 {
				m.viewportStartIndex = min(m.viewportStartIndex+1, len(m.issues)-(m.height-START_ROW-1))
			}
			return m, nil
		case "k":
			m.cursorIndex = max(m.cursorIndex-1, 0)
			if m.cursorIndex < m.viewportStartIndex {
				m.viewportStartIndex = m.cursorIndex
			}
			return m, nil
		case "h":
			m.cursorIndex = 0
			m.viewportStartIndex = 0
			m.selectedIssueIndex = -1
			if m.currentPage == 1 {
				return m, nil
			}
			m.currentPage--
			err := m.fetchIssues()
			checkErr(err)
			return m, nil
		case "l":
			m.cursorIndex = 0
			m.viewportStartIndex = 0
			m.selectedIssueIndex = -1
			if m.currentPage == m.lastPage {
				return m, nil
			}
			m.currentPage++
			err := m.fetchIssues()
			checkErr(err)
			return m, nil
		case "g":
			m.cursorIndex = 0
			m.viewportStartIndex = 0
			return m, nil
		case "G":
			m.cursorIndex = len(m.issues) - 1
			m.viewportStartIndex = len(m.issues) - m.height + START_ROW + 1
			return m, nil
		case "H":
			m.cursorIndex = m.viewportStartIndex
			return m, nil
		case "L":
			m.cursorIndex = min(m.viewportStartIndex+m.height-START_ROW-2, len(m.issues)-1)
			return m, nil
		case "enter":
			m.selectedIssueIndex = m.cursorIndex
			return m, nil
		case "esc":
			m.selectedIssueIndex = -1
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	maxRow := m.height - START_ROW - 1

	tableWidth := m.width
	if m.selectedIssueIndex != -1 {
		tableWidth = m.width / 2
	}

	headerStyle := issueListItemStyle(m.width)
	headerContent := fmt.Sprintf(`
Repo: %s\tPage: %d

`, m.repo, m.currentPage)
	header := headerStyle.Render(headerContent)

	list := strings.Builder{}
	for i := 0; i < maxRow; i++ {
		if i+m.viewportStartIndex >= len(m.issues) {
			style := issueListItemStyle(tableWidth)
			// Fill in blank space
			for j := i; j < maxRow; j++ {
				list.WriteString(style.Render("\n"))
			}
			break
		}

		issue := m.issues[i+m.viewportStartIndex]
		var listStyle lipgloss.Style
		if i+m.viewportStartIndex == m.cursorIndex {
			listStyle = selectedIssueListItemStyle(tableWidth)
		} else {
			listStyle = issueListItemStyle(tableWidth)
		}

		issueString := strconv.Itoa(*issue.Number) + " " + *issue.Title
		if len(issueString) >= tableWidth {
			issueString = issueString[:tableWidth-3] + "..."
		}
		list.WriteString(listStyle.Render(issueString) + "\n")
	}

	selectedIssueContent := ""
	if m.selectedIssueIndex != -1 {
		selectedIssue := m.issues[m.selectedIssueIndex]
		style := lipgloss.NewStyle().Width(tableWidth).Height(m.height-START_ROW-2).Background(lipgloss.Color("#0F0")).Foreground(lipgloss.Color("#000")).Border(lipgloss.NormalBorder(), true)
		lines := strings.Split(*selectedIssue.Body, "\n")
		lines = lines[:min(len(lines), m.height-START_ROW-4)]
		for i := 0; i < len(lines); i++ {
			if len(lines[i]) >= tableWidth-1 {
				lines[i] = lines[i][:tableWidth-1]
			}
		}

		body := strings.Join(lines, "\n")
		selectedIssueContent = style.Render(body)
	}

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			list.String(),
			selectedIssueContent,
		))

	return view
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
	m.lastPage = response.LastPage
	return nil
}

func issueListItemStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Background(lipgloss.Color("#000")).Foreground(lipgloss.Color("#FFF"))
}

func selectedIssueListItemStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Background(lipgloss.Color("#FFF")).Foreground(lipgloss.Color("#000"))
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
