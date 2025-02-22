package pages

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/gtui/internal/components"
)

type State int

const (
	LoadingState State = iota
	ReadyState

	startRow    = 2
	borderWidth = 1
)

var (
	selectedIssueGlamourStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("62")).
					PaddingRight(2)

	borderedPageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center)
)

type IssuesReadyMsg struct {
	issues         []*github.Issue
	lastIssuesPage int
}

type IssuesPageComponents struct {
	spinner    spinner.Model
	issuesList components.IssuesModel
}

type IssuesPageModel struct {
	width  int
	height int

	state             State
	currentIssuesPage int
	lastIssuesPage    int

	client *github.Client
	repo   string
	issues []*github.Issue

	selectedIssue         *github.Issue
	selectedIssueViewport viewport.Model
	selectedIssueRenderer *glamour.TermRenderer

	components IssuesPageComponents
}

func NewIssuesPage(client *github.Client, repo string, width int, height int) IssuesPageModel {
	m := IssuesPageModel{
		state:             LoadingState,
		client:            client,
		repo:              repo,
		width:             width,
		height:            height,
		currentIssuesPage: 1,
		components: IssuesPageComponents{
			spinner: spinner.New(
				spinner.WithSpinner(spinner.Dot),
				spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
			),
			issuesList: components.NewIssuesModel(width, height-1),
		},
	}
	m.selectedIssueViewport = viewport.New(m.width/2-borderWidth*2, m.height-startRow-1)
	m.selectedIssueViewport.Style = selectedIssueGlamourStyle
	glamourRenderWidth := m.width/2 - m.selectedIssueViewport.Style.GetHorizontalFrameSize() - borderWidth*2
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	m.selectedIssueRenderer = renderer
	checkErr(err)

	return m
}

func (m IssuesPageModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchIssues(0),
		m.components.spinner.Tick,
	)
}

func (m IssuesPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "j":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "k":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "g":
			if m.selectedIssue != nil {
				m.selectedIssueViewport.GotoTop()
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "G":
			if m.selectedIssue != nil {
				m.selectedIssueViewport.GotoBottom()
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "H":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "L":
			if m.selectedIssue != nil {
				m.selectedIssueViewport, _ = m.selectedIssueViewport.Update(msg)
				return m, nil
			}

			m.components.issuesList, _ = m.components.issuesList.Update(msg)
			return m, nil
		case "enter":
			m.selectedIssue = m.components.issuesList.GetSelectedIssue()
			m.components.issuesList.SetWidth(m.width / 2)
			str, err := m.selectedIssueRenderer.Render(*m.selectedIssue.Body)
			checkErr(err)
			m.selectedIssueViewport.SetContent(str)
			return m, nil
		case "esc":
			m.selectedIssue = nil
			m.components.issuesList.SetWidth(m.width)
			return m, nil
		case "[":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.issuesList.ResetViewport()
			m.currentIssuesPage--
			return m, m.fetchIssues(m.currentIssuesPage)
		case "]":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.issuesList.ResetViewport()
			m.currentIssuesPage++
			return m, m.fetchIssues(m.currentIssuesPage)
		case "{":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.issuesList.ResetViewport()
			m.currentIssuesPage = 1
			return m, m.fetchIssues(m.currentIssuesPage)
		case "}":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.issuesList.ResetViewport()
			m.currentIssuesPage = m.lastIssuesPage
			return m, m.fetchIssues(m.currentIssuesPage)
		}
	case IssuesReadyMsg:
		m.lastIssuesPage = msg.lastIssuesPage
		m.issues = msg.issues
		m.state = ReadyState
		m.components.issuesList.SetIssues(m.issues)
		return m, nil
	default:
		var cmd tea.Cmd
		m.components.spinner, cmd = m.components.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m IssuesPageModel) View() string {
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		m.header(),
		m.body(),
	)

	return view
}

func (m IssuesPageModel) header() string {
	return m.repo
}

func (m IssuesPageModel) body() string {
	switch m.state {
	case LoadingState:
		return borderedPageStyle.
			Width(m.width - 2).
			Height(m.height - startRow - 1).
			Render(fmt.Sprintf("%s Loading Issues from %s", m.components.spinner.View(), m.repo))
	case ReadyState:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.components.issuesList.View(),
			m.selectedIssueComponent(),
		)
	default:
		return ""
	}
}

func (m IssuesPageModel) selectedIssueComponent() string {
	if m.selectedIssue != nil {
		return m.selectedIssueViewport.View()
	}
	return ""
}

func (m *IssuesPageModel) fetchIssues(page int) tea.Cmd {
	m.state = LoadingState

	return func() tea.Msg {
		searchString := fmt.Sprintf("repo:%s is:open is:issue", m.repo)
		result, response, err := m.client.Search.Issues(context.Background(), searchString, &github.SearchOptions{
			Sort:        "created",
			Order:       "desc",
			ListOptions: github.ListOptions{Page: page, PerPage: 50},
		})

		// TODO: send err as message
		checkErr(err)

		return IssuesReadyMsg{
			issues:         result.Issues,
			lastIssuesPage: response.LastPage,
		}
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
