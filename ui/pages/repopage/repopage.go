package repopage

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/ghtui/ui/components"
	"github.com/alex-laycalvert/ghtui/utils"
)

type RepoPageModel struct {
	id     string
	width  int
	height int

	isLoaded bool
	state    utils.ComponentState
	repo     string
	client   *github.Client

	componentGroup          utils.ComponentGroup
	spinnerComponent        string
	markdownViewerComponent string
}

type repoLoadingMsg struct{}

type repoReadyMsg struct {
	content string
}

func NewRepoPage(id string, client *github.Client, repo string, width int, height int) RepoPageModel {
	spinner := components.NewSpinnerComponent()
	markdownViewer := components.NewMarkdownViewerComponent(
		width,
		height,
		lipgloss.NewStyle(),
	)

	return RepoPageModel{
		id:                      id,
		isLoaded:                false,
		client:                  client,
		repo:                    repo,
		width:                   width,
		height:                  height,
		componentGroup:          utils.NewComponentGroup(spinner, markdownViewer),
		spinnerComponent:        spinner.ID(),
		markdownViewerComponent: markdownViewer.ID(),
	}
}

func (m RepoPageModel) ID() string {
	return m.id
}

func (m RepoPageModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchRepo(),
		m.componentGroup.Init(),
	)
}

func (m RepoPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case utils.FocusMsg:
		if m.id != msg.ID {
			return m, m.componentGroup.UpdateFocused(msg)
		}
		if m.isLoaded {
			return m, nil
		}
		return m, m.fetchRepo()
	case utils.BlurMsg:
		if m.id != msg.ID {
			return m, m.componentGroup.UpdateFocused(msg)
		}
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		default:
			cmd := m.componentGroup.UpdateFocused(msg)
			return m, cmd
		}
	case repoReadyMsg:
		m.state = utils.ReadyState
		m.isLoaded = true
		return m, tea.Batch(
			m.componentGroup.Update(m.markdownViewerComponent, components.MarkdownViewerSetContentMsg{
				Content: msg.content,
			}),
			m.componentGroup.FocusOn(m.markdownViewerComponent),
		)
	case repoLoadingMsg:
		m.state = utils.LoadingState
		return m, m.componentGroup.FocusOn(m.spinnerComponent)
	default:
		cmd := m.componentGroup.Update(m.spinnerComponent, msg)
		return m, cmd
	}
}

func (m RepoPageModel) View() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(m.body())
}

func (m RepoPageModel) body() string {
	switch m.state {
	case utils.LoadingState:
		return fmt.Sprintf(
			"%s Loading Repo %s",
			m.componentGroup.GetComponent(m.spinnerComponent).View(),
			m.repo,
		)
	case utils.ReadyState:
		return m.componentGroup.GetComponent(m.markdownViewerComponent).View()
	default:
		return ""
	}
}

func (m *RepoPageModel) fetchRepo() tea.Cmd {
	return tea.Sequence(
		repoLoadingCmd,
		func() tea.Msg {
			parts := strings.Split(m.repo, "/")
			owner := parts[0]
			repoName := parts[1]
			// TODO: handle err
			content, _, _ := m.client.Repositories.GetReadme(context.Background(), owner, repoName, nil)
			markdown, _ := content.GetContent()

			return repoReadyMsg{content: markdown}
		},
	)
}

func repoLoadingCmd() tea.Msg {
	return repoLoadingMsg{}
}
