package issues

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/ghtui/internal/components"
	"github.com/alex-laycalvert/ghtui/internal/utils"
)

const (
	spinnerComponent        components.ComponentName = "spinner"
	issuesListComponent     components.ComponentName = "issuesList"
	markdownViewerComponent components.ComponentName = "markdownViewer"
)

type IssuesPageModel struct {
	width  int
	height int

	repo              string
	client            *github.Client
	state             utils.ComponentState
	currentIssuesPage int
	lastIssuesPage    int
	selectedIssue     *github.Issue

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
			components.NameComponent(
				spinnerComponent,
				components.NewSpinnerComponent(),
			),
			components.NameComponent(
				issuesListComponent,
				components.NewIssuesListComponent(width, height),
			),
			components.NameComponent(
				markdownViewerComponent,
				components.NewMarkdownViewerComponent(
					width/2,
					height,
					lipgloss.NewStyle().
						Border(lipgloss.NormalBorder()).
						BorderForeground(lipgloss.Color("62")).
						UnsetBorderTop().
						UnsetBorderRight().
						UnsetBorderBottom().
						PaddingRight(2),
				),
			),
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
			m.selectedIssue = issue
			m.components.Update(issuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width / 2,
			})
			m.components.Update(markdownViewerComponent, components.MarkdownViewerSetContentMsg{
				Content: *issue.Body,
			})
			m.components.FocusOn(markdownViewerComponent)
			return m, nil
		case "esc":
			m.selectedIssue = nil
			m.components.Update(issuesListComponent, components.IssuesListUpdateWidthMsg{
				Width: m.width,
			})
			m.components.FocusOn(issuesListComponent)
			return m, nil
		case "[":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage--
			return m, m.fetchIssues(m.currentIssuesPage)
		case "]":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage >= m.lastIssuesPage {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage++
			return m, m.fetchIssues(m.currentIssuesPage)
		case "{":
			if m.selectedIssue != nil {
				return m, nil
			}

			if m.currentIssuesPage == 1 {
				return m, nil
			}
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = 1
			return m, m.fetchIssues(m.currentIssuesPage)
		case "}":
			if m.selectedIssue != nil {
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
		if m.selectedIssue != nil {
			return lipgloss.JoinHorizontal(
				lipgloss.Top,
				m.components.GetComponent(issuesListComponent).View(),
				m.components.GetComponent(markdownViewerComponent).View(),
			)
		}
		return m.components.GetComponent(issuesListComponent).View()
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

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
