package repo

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"

	"github.com/alex-laycalvert/gtui/internal/components"
	"github.com/alex-laycalvert/gtui/internal/utils"
)

const (
	spinnerComponent components.ComponentName = "spinner"
)

type RepoPageModel struct {
	width  int
	height int

	state  utils.ComponentState
	repo   string
	client *github.Client

	components components.ComponentGroup
}

type RepoReadyMsg struct {
	content string
}

func NewRepoPage(client *github.Client, repo string, width int, height int) RepoPageModel {
	components := components.NewComponentGroup(
		components.NameComponent(spinnerComponent, components.NewSpinnerComponent()),
	)
	return RepoPageModel{
		client:     client,
		repo:       repo,
		width:      width,
		height:     height,
		components: components,
	}
}

func (m RepoPageModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchRepo(),
		m.components.Init(),
	)
}

func (m RepoPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		default:
			cmd := m.components.UpdateFocused(msg)
			return m, cmd
		}
	case RepoReadyMsg:
		m.state = utils.ReadyState
		return m, nil
	default:
		cmd := m.components.Update(spinnerComponent, msg)
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
			m.components.GetComponent(spinnerComponent).View(),
			m.repo,
		)
	case utils.ReadyState:
		return "Hello, Repo!"
	default:
		return ""
	}
}

func (m *RepoPageModel) fetchRepo() tea.Cmd {
	m.state = utils.LoadingState

	return func() tea.Msg {
		parts := strings.Split(m.repo, "/")
		owner := parts[0]
		repoName := parts[1]
		// TODO: handle err
		content, _, _ := m.client.Repositories.GetReadme(context.Background(), owner, repoName, nil)

		return RepoReadyMsg{
			content: *content.Content,
		}
	}
}
