package app

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v69/github"
	"golang.org/x/term"

	"github.com/alex-laycalvert/ghtui/ui/components"
	"github.com/alex-laycalvert/ghtui/ui/pages/issuespage"
	"github.com/alex-laycalvert/ghtui/ui/pages/repopage"
)

type App struct {
	model appModel
}

func New(token string, repoName string) (*App, error) {
	client := github.NewClient(nil).WithAuthToken(token)

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}

	pageWidth := width - 6
	pageHeight := height - 6

	repo := repopage.NewRepoPage("Repo", client, repoName, pageWidth, pageHeight)
	issues := issuespage.NewIssuesPage("Issues", client, repoName, pageWidth, pageHeight)

	model := appModel{
		client: client,
		repo:   repoName,
		width:  width,
		height: height,

		pageGroup: components.NewComponentGroup(
			repo,
			issues,
		),
		issuesPage: issues.ID(),
		repoPage:   repo.ID(),
	}

	return &App{model: model}, nil
}

func (app *App) Run() error {
	if _, err := tea.NewProgram(app.model).Run(); err != nil {
		return err
	}
	return nil
}

type appModel struct {
	width  int
	height int

	client *github.Client
	repo   string

	pageGroup  components.ComponentGroup
	issuesPage string
	repoPage   string
}

func (model appModel) Init() tea.Cmd {
	return model.pageGroup.Init()
}

func (model appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return model, tea.Quit
		case "tab":
			return model, model.pageGroup.FocusNext()
		case "shift+tab":
			return model, model.pageGroup.FocusPrevious()
		default:
			return model, model.pageGroup.UpdateFocused(msg)
		}
	default:
		return model, model.pageGroup.UpdateFocused(msg)
	}
}

func tabBorderStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(highlightColor).
		Padding(0, 1)
	return style
}

var (
	docStyle         = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle = tabBorderStyle()
	activeTabStyle   = inactiveTabStyle.
				Bold(true)
	windowStyle = lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Border(lipgloss.RoundedBorder())
)

func (model appModel) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	pages := model.pageGroup.GetComponents()
	currentPage := model.pageGroup.GetFocusedComponent()
	for _, t := range pages {
		var style lipgloss.Style
		isActive := t.ID() == currentPage.ID()
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(string(t.ID())))
	}

	header := lipgloss.NewStyle().
		MarginLeft(1).
		Padding(1).
		Render(model.repo)
	row := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...),
		header,
	)
	doc.WriteString(row + "\n")
	doc.WriteString(
		windowStyle.
			Render(currentPage.View()),
	)
	return docStyle.
		Width(model.width).
		Height(model.height).
		Render(doc.String())
}
