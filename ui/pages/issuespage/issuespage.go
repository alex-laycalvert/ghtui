package issuespage

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/ghtui/ui/components"
	"github.com/alex-laycalvert/ghtui/utils"
)

type IssuesPageModel struct {
	id     string
	width  int
	height int

	repo              string
	client            *github.Client
	state             utils.ComponentState
	currentIssuesPage int
	lastIssuesPage    int
	selectedIssue     *github.Issue
	search            string

	componentGroup          components.ComponentGroup
	spinnerComponent        string
	issuesListComponent     string
	markdownViewerComponent string
	textInputComponent      string
}

type issuesLoadingMsg struct{}

type issuesReadyMsg struct {
	issues         []*github.Issue
	lastIssuesPage int
}

func NewIssuesPage(id string, client *github.Client, repo string, width int, height int) IssuesPageModel {
	spinner := components.NewSpinnerComponent()
	issuesList := components.NewIssuesListComponent(width, height)
	markdownViewer := components.NewMarkdownViewerComponent(
		width/2,
		height,
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			UnsetBorderTop().
			UnsetBorderRight().
			UnsetBorderBottom().
			PaddingRight(2),
	)
	textInput := components.NewTextInputComponent("Search", width)

	m := IssuesPageModel{
		id:                id,
		state:             utils.LoadingState,
		client:            client,
		repo:              repo,
		width:             width,
		height:            height,
		currentIssuesPage: 1,
		componentGroup: components.NewComponentGroup(
			spinner,
			issuesList,
			markdownViewer,
			textInput,
		),
		spinnerComponent:        spinner.ID(),
		issuesListComponent:     issuesList.ID(),
		markdownViewerComponent: markdownViewer.ID(),
		textInputComponent:      textInput.ID(),
	}

	return m
}

func (m IssuesPageModel) ID() string {
	return m.id
}

func (m IssuesPageModel) Init() tea.Cmd {
	return tea.Sequence(
		m.fetchIssues("", 0),
		m.componentGroup.Init(),
	)
}

func (m IssuesPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case utils.FocusMsg:
		if m.id != msg.ID {
			return m, m.componentGroup.UpdateFocused(msg)
		}
		return m, m.fetchIssues(m.search, m.currentIssuesPage)
	case utils.BlurMsg:
		if m.id != msg.ID {
			return m, m.componentGroup.UpdateFocused(msg)
		}
		return m, nil
	case tea.KeyMsg:
		switch k := msg.String(); {
		case k == "enter" && m.state == utils.ReadyState && m.componentGroup.IsFocused(m.issuesListComponent):
			issue := m.getSelectedIssue()
			m.selectedIssue = issue
			return m, tea.Sequence(
				m.componentGroup.Update(m.issuesListComponent, components.ComponentUpdateSizeMsg{
					Width: m.width / 2,
				}),
				m.componentGroup.Update(m.textInputComponent, components.ComponentUpdateSizeMsg{
					Width: m.width / 2,
				}),
				m.componentGroup.Update(m.markdownViewerComponent, components.MarkdownViewerSetContentMsg{
					Content: *issue.Body,
				}),
				m.componentGroup.FocusOn(m.markdownViewerComponent),
			)
		case k == "esc":
			if m.componentGroup.IsFocused(m.textInputComponent) {
				return m, tea.Sequence(
					m.componentGroup.Update(m.issuesListComponent, components.ComponentUpdateSizeMsg{
						Height: m.height,
					}),
					m.componentGroup.FocusOn(m.issuesListComponent),
				)
			} else if m.componentGroup.IsFocused(m.markdownViewerComponent) {
				m.selectedIssue = nil
				return m, tea.Sequence(
					m.componentGroup.Update(m.issuesListComponent, components.ComponentUpdateSizeMsg{
						Width: m.width,
					}),
					m.componentGroup.FocusOn(m.issuesListComponent),
				)
			} else {
				cmds := []tea.Cmd{
					m.componentGroup.Update(m.textInputComponent, components.TextInputClearMsg{}),
				}
				if m.search != "" {
					m.search = ""
					cmds = append(cmds, m.fetchIssues("", 0))
				}
				return m, tea.Sequence(cmds...)
			}
		case k == "[" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage > 1:
			m.componentGroup.Update(m.issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage--
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "]" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage < m.lastIssuesPage:
			m.componentGroup.Update(m.issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage++
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "{" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage > 1:
			m.componentGroup.Update(m.issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = 1
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "}" && m.state == utils.ReadyState && m.selectedIssue == nil && m.currentIssuesPage < m.lastIssuesPage:
			m.componentGroup.Update(m.issuesListComponent, components.IssuesListResetViewportMsg{})
			m.currentIssuesPage = m.lastIssuesPage
			return m, m.fetchIssues("", m.currentIssuesPage)
		case k == "/" && m.state == utils.ReadyState && !m.componentGroup.IsFocused(m.textInputComponent):
			return m, tea.Sequence(
				m.componentGroup.Update(m.issuesListComponent, components.ComponentUpdateSizeMsg{
					Height: m.height - 1,
				}),
				m.componentGroup.FocusOn(m.textInputComponent),
			)
		default:
			if m.state == utils.LoadingState {
				return m, nil
			}
			return m, m.componentGroup.UpdateFocused(msg)
		}
	case components.TextInputSubmitMsg:
		cmds := []tea.Cmd{m.componentGroup.FocusOn(m.issuesListComponent)}
		if m.search != msg.Value {
			m.search = msg.Value
			cmds = append(cmds, m.fetchIssues(msg.Value, 0))
		}
		return m, tea.Sequence(cmds...)
	case issuesReadyMsg:
		m.lastIssuesPage = msg.lastIssuesPage
		m.state = utils.ReadyState
		return m, tea.Batch(
			m.componentGroup.FocusOn(m.issuesListComponent),
			m.componentGroup.Update(m.issuesListComponent, components.IssuesListUpdateIssuesMsg{
				Issues: msg.issues,
			}),
		)
	case issuesLoadingMsg:
		m.state = utils.LoadingState
		return m, m.componentGroup.FocusOn(m.spinnerComponent)
	default:
		return m, m.componentGroup.UpdateFocused(msg)
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
				m.componentGroup.GetComponent(m.spinnerComponent).View(),
				m.repo,
			))
	case utils.ReadyState:
		issuesList := m.componentGroup.GetComponent(m.issuesListComponent).View()
		if m.componentGroup.IsFocused(m.textInputComponent) || m.search != "" {
			issuesList = lipgloss.JoinVertical(
				lipgloss.Left,
				issuesList,
				m.componentGroup.GetComponent(m.textInputComponent).View(),
			)
		}

		if m.selectedIssue == nil {
			return issuesList
		}

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			issuesList,
			m.componentGroup.GetComponent(m.markdownViewerComponent).View(),
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
	return m.componentGroup.
		GetComponent(m.issuesListComponent).(components.IssuesListModel).
		GetSelectedIssue()
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
