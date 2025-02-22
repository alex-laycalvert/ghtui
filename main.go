package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v69/github"
	"golang.org/x/term"

	"github.com/alex-laycalvert/gtui/internal/pages"
)

type model struct {
	client *github.Client

	width, height int

	pages            []tea.Model
	currentPageIndex int
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

	issuesPage := pages.NewIssuesPage(client, repo, width, height)

	m := model{
		client: client,
		width:  width,
		height: height,

		currentPageIndex: 0,
		pages:            []tea.Model{issuesPage},
	}

	_, err = tea.NewProgram(m).Run()
	checkErr(err)
}

func (m model) Init() tea.Cmd {
	initCmds := make([]tea.Cmd, len(m.pages))
	for i, page := range m.pages {
		initCmds[i] = page.Init()
	}
	return tea.Batch(initCmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.pages[m.currentPageIndex], cmd = m.pages[m.currentPageIndex].Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.pages[m.currentPageIndex], cmd = m.pages[m.currentPageIndex].Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	currentPage := m.pages[m.currentPageIndex]
	return currentPage.View()
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
