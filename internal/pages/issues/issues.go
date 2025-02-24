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
	textInputComponent      components.ComponentName = "textInput"
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
	search            string

	components components.ComponentGroup
}

type issuesLoadingMsg struct{}

type issuesReadyMsg struct {
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
			components.NameComponent(
				textInputComponent,
				components.NewTextInputComponent("Search", width),
			),
		),
	}

	return m
}

func (m IssuesPageModel) Init() tea.Cmd {
	return tea.Sequence(
		m.fetchIssues("", 0),
		m.components.Init(),
	)
}

func (m IssuesPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch k := msg.String(); {
		case k == "enter" && m.state == utils.ReadyState && m.components.IsFocused(issuesListComponent):
			issue := m.getSelectedIssue()
			m.selectedIssue = issue
			return m, tea.Sequence(
				m.components.Update(issuesListComponent, components.ComponentUpdateSizeMsg{
					Width: m.width / 2,
				}),
				m.components.Update(textInputComponent, components.ComponentUpdateSizeMsg{
					Width: m.width / 2,
				}),
				m.components.Update(markdownViewerComponent, components.MarkdownViewerSetContentMsg{
					Content: *issue.Body,
				}),
				m.components.FocusOn(markdownViewerComponent),
			)
		case k == "esc":
			if m.components.IsFocused(textInputComponent) {
				return m, tea.Sequence(
					m.components.Update(issuesListComponent, components.ComponentUpdateSizeMsg{
						Height: m.height,
					}),
					m.components.FocusOn(issuesListComponent),
				)
			} else if m.components.IsFocused(markdownViewerComponent) {
				m.selectedIssue = nil
				return m, tea.Sequence(
					m.components.Update(issuesListComponent, components.ComponentUpdateSizeMsg{
						Width: m.width,
					}),
					m.components.FocusOn(issuesListComponent),
				)
			} else {
				cmds := []tea.Cmd{
					m.components.Update(textInputComponent, components.TextInputClearMsg{}),
				}
				if m.search != "" {
					m.search = ""
					cmds = append(cmds, m.fetchIssues("", 0))
				}
				return m, tea.Sequence(cmds...)
			}
		case k == "[" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage > 1:
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage--
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "]" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage < m.lastIssuesPage:
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage++
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "{" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage > 1:
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = 1
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "}" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage < m.lastIssuesPage:
			m.components.Update(issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = m.lastIssuesPage
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "/" && m.state == utils.ReadyState && !m.components.IsFocused(textInputComponent):
			return m, tea.Sequence(
				m.components.Update(issuesListComponent, components.ComponentUpdateSizeMsg{
					Height: m.height - 1,
				}),
				m.components.FocusOn(textInputComponent),
			)
		default:
			if m.state == utils.LoadingState {
				return m, nil
			}
			return m, m.components.UpdateFocused(msg)
		}
	case components.TextInputSubmitMsg:
		cmds := []tea.Cmd{m.components.FocusOn(issuesListComponent)}
		if m.search != msg.Value {
			m.search = msg.Value
			cmds = append(cmds, m.fetchIssues(msg.Value, 0))
		}
		return m, tea.Sequence(cmds...)
	case issuesReadyMsg:
		m.lastIssuesPage = msg.lastIssuesPage
		m.state = utils.ReadyState
		return m, tea.Batch(
			m.components.FocusOn(issuesListComponent),
			m.components.Update(issuesListComponent, components.IssuesListUpdateIssuesMsg{
				Issues: msg.issues,
			}),
		)
	case issuesLoadingMsg:
		m.state = utils.LoadingState
		return m, m.components.FocusOn(spinnerComponent)
	default:
		return m, m.components.UpdateFocused(msg)
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
		issuesList := m.components.GetComponent(issuesListComponent).View()
		if m.components.IsFocused(textInputComponent) || m.search != "" {
			issuesList = lipgloss.JoinVertical(
				lipgloss.Left,
				issuesList,
				m.components.GetComponent(textInputComponent).View(),
			)
		}

		if m.selectedIssue == nil {
			return issuesList
		}

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			issuesList,
			m.components.GetComponent(markdownViewerComponent).View(),
		)
	default:
		return ""
	}
}

func (m *IssuesPageModel) fetchIssues(searchTerm string, page int) tea.Cmd {
	return tea.Sequence(
		issuesLoadingCmd,
		func() tea.Msg {
			searchString := fmt.Sprintf("repo:%s is:open is:issue %s", m.repo, searchTerm)
			result, response, err := m.client.Search.Issues(context.Background(), searchString, &github.SearchOptions{
				Sort:        "created",
				Order:       "desc",
				ListOptions: github.ListOptions{Page: page, PerPage: 50},
			})

			// TODO: send err as message
			checkErr(err)

			return issuesReadyMsg{
				issues:         result.Issues,
				lastIssuesPage: response.LastPage,
			}
		},
	)
}

func issuesLoadingCmd() tea.Msg {
	return issuesLoadingMsg{}
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
