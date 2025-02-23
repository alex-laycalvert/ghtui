package pages

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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

var borderedPageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	AlignHorizontal(lipgloss.Center).
	AlignVertical(lipgloss.Center)

type IssuesReadyMsg struct {
	issues         []*github.Issue
	lastIssuesPage int
}

const (
	SpinnerComponent       components.ComponentName = "spinner"
	SelectedIssueComponent components.ComponentName = "selectedIssue"
	IssuesListComponent    components.ComponentName = "issuesList"
)

type IssuesPageComponents struct {
	spinner       spinner.Model
	selectedIssue components.IssueModel
	issuesList    components.IssuesListModel
}

type IssuesPageModel struct {
	width  int
	height int

	repo              string
	client            *github.Client
	state             State
	currentIssuesPage int
	lastIssuesPage    int

	components components.ComponentGroup
}

func NewIssuesPage(client *github.Client, repo string, width int, height int) IssuesPageModel {
	selectedIssue, err := components.NewIssueComponent(nil, width/2, height-1)
	checkErr(err)

	m := IssuesPageModel{
		state:             LoadingState,
		client:            client,
		repo:              repo,
		width:             width,
		height:            height,
		currentIssuesPage: 1,
		components: components.NewComponentGroup([]components.Component{
			components.NameComponent(SelectedIssueComponent, selectedIssue),
			components.NameComponent(SpinnerComponent, components.NewSpinnerComponent()),
			components.NameComponent(IssuesListComponent, components.NewIssuesListComponent(width, height-1)),
		}),
	}
	checkErr(err)

	return m
}

func (m IssuesPageModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchIssues(0),
		m.components.Init(),
	)
}

func (m IssuesPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			issue := m.getSelectedIssue()
			m.components.Update(IssuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width / 2,
			})
			m.components.Update(SelectedIssueComponent, components.IssueSetIssueMsg{
				Issue: issue,
			})
			return m, nil
		case "esc":
			m.components.Update(IssuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width,
			})
			m.components.Update(SelectedIssueComponent, components.IssueSetIssueMsg{
				Issue: nil,
			})
			return m, nil
		case "[":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(IssuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage--
			return m, m.fetchIssues(m.currentIssuesPage)
		case "]":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.Update(IssuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage++
			return m, m.fetchIssues(m.currentIssuesPage)
		case "{":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(IssuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = 1
			return m, m.fetchIssues(m.currentIssuesPage)
		case "}":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.Update(IssuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = m.lastIssuesPage
			return m, m.fetchIssues(m.currentIssuesPage)
		default:
			cmd := m.components.UpdateFocused(msg)
			return m, cmd
		}
	case IssuesReadyMsg:
		m.lastIssuesPage = msg.lastIssuesPage
		m.state = ReadyState
		m.components.Update(IssuesListComponent, components.IssuesListUpdateIssuesMsg{
			Issues: msg.issues,
		})
		m.components.FocusOn(IssuesListComponent)
		return m, nil
	default:
		cmd := m.components.Update(SpinnerComponent, msg)
		return m, cmd
	}
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
			Render(fmt.Sprintf("%s Loading Issues from %s", m.components.GetComponent(SpinnerComponent).View(), m.repo))
	case ReadyState:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.components.GetComponent(IssuesListComponent).View(),
			m.components.GetComponent(SelectedIssueComponent).View(),
		)
	default:
		return ""
	}
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

// TODO: maybe make this a msg that is sent to this component?
func (m IssuesPageModel) getSelectedIssue() *github.Issue {
	return m.components.
		GetComponent(IssuesListComponent).(components.NamedComponent[components.IssuesListModel]).
		Component.GetSelectedIssue()
}

// TODO: maybe make this a msg that is sent to this component?
func (m IssuesPageModel) getDisplayedIssue() *github.Issue {
	return m.components.
		GetComponent(SelectedIssueComponent).(components.NamedComponent[components.IssueModel]).
		Component.GetIssue()
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
