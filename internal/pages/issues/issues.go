package issues

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/gtui/internal/components"
	"github.com/alex-laycalvert/gtui/internal/utils"
)

const (
	spinnerComponent       components.ComponentName = "spinner"
	selectedIssueComponent components.ComponentName = "selectedIssue"
	issuesListComponent    components.ComponentName = "issuesList"
)

type IssuesPageModel struct {
	width  int
	height int

	repo              string
	client            *github.Client
	state             utils.ComponentState
	currentIssuesPage int
	lastIssuesPage    int

	components components.ComponentGroup
}

type IssuesReadyMsg struct {
	issues         []*github.Issue
	lastIssuesPage int
}

func NewIssuesPage(client *github.Client, repo string, width int, height int) IssuesPageModel {
	m := IssuesPageModel{
		state:             utils.LoadingState,
		client:            client,
		repo:              repo,
		width:             width,
		height:            height,
		currentIssuesPage: 1,
		components: components.NewComponentGroup(
			components.NameComponent(selectedIssueComponent, components.NewIssueComponent(nil, width/2, height)),
			components.NameComponent(spinnerComponent, components.NewSpinnerComponent()),
			components.NameComponent(issuesListComponent, components.NewIssuesListComponent(width, height)),
		),
	}

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
			if m.state == utils.LoadingState {
				return m, nil
			}
			issue := m.getSelectedIssue()
			m.components.Update(issuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width / 2,
			})
			m.components.Update(selectedIssueComponent, components.IssueSetIssueMsg{
				Issue: issue,
			})
			m.components.FocusOn(selectedIssueComponent)
			return m, nil
		case "esc":
			m.components.Update(issuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width,
			})
			m.components.Update(selectedIssueComponent, components.IssueSetIssueMsg{
				Issue: nil,
			})
			m.components.FocusOn(issuesListComponent)
			return m, nil
		case "[":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage--
			return m, m.fetchIssues(m.currentIssuesPage)
		case "]":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage++
			return m, m.fetchIssues(m.currentIssuesPage)
		case "{":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = 1
			return m, m.fetchIssues(m.currentIssuesPage)
		case "}":
			if m.getDisplayedIssue() != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = m.lastIssuesPage
			return m, m.fetchIssues(m.currentIssuesPage)
		default:
			cmd := m.components.UpdateFocused(msg)
			return m, cmd
		}
	case IssuesReadyMsg:
		m.lastIssuesPage = msg.lastIssuesPage
		m.state = utils.ReadyState
		m.components.Update(issuesListComponent, components.IssuesListUpdateIssuesMsg{
			Issues: msg.issues,
		})
		m.components.FocusOn(issuesListComponent)
		return m, nil
	default:
		cmd := m.components.Update(spinnerComponent, msg)
		return m, cmd
	}
}

func (m IssuesPageModel) View() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(m.body())
}

func (m IssuesPageModel) body() string {
	switch m.state {
	case utils.LoadingState:
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(fmt.Sprintf(
				"%s Loading Issues from %s",
				m.components.GetComponent(spinnerComponent).View(),
				m.repo,
			))
	case utils.ReadyState:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.components.GetComponent(issuesListComponent).View(),
			m.components.GetComponent(selectedIssueComponent).View(),
		)

	default:
		return ""
	}
}

func (m *IssuesPageModel) fetchIssues(page int) tea.Cmd {
	m.state = utils.LoadingState

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
		GetComponent(issuesListComponent).(components.NamedComponent[components.IssuesListModel]).
		Component.GetSelectedIssue()
}

// TODO: maybe make this a msg that is sent to this component?
func (m IssuesPageModel) getDisplayedIssue() *github.Issue {
	return m.components.
		GetComponent(selectedIssueComponent).(components.NamedComponent[components.IssueModel]).
		Component.GetIssue()
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
