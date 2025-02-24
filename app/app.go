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

const (
	repoPageComponent   components.ComponentName = "Repo"
	issuesPageComponent components.ComponentName = "Issues"
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

	pages := components.NewComponentGroup(
		components.NameComponent(
			repoPageComponent,
			repopage.NewRepoPage(client, repoName, pageWidth, pageHeight),
		),
		components.NameComponent(
			issuesPageComponent,
			issuespage.NewIssuesPage(client, repoName, pageWidth, pageHeight),
		),
	)

	model := appModel{
		client: client,
		repo:   repoName,
		width:  width,
		height: height,
		pages:  pages,
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

	pages components.ComponentGroup
}

func (model appModel) Init() tea.Cmd {
	return model.pages.Init()
}

func (model appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return model, tea.Quit
		case "tab":
			model.pages.FocusNext()
			return model, model.pages.GetFocusedComponent().Init()
		case "shift+tab":
			model.pages.FocusPrevious()
			return model, model.pages.GetFocusedComponent().Init()
		default:
			return model, model.pages.UpdateFocused(msg)
		}
	default:
		return model, model.pages.UpdateFocused(msg)
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

	pages := model.pages.GetComponents()
	currentPage := model.pages.GetFocusedComponent()
	for _, t := range pages {
		var style lipgloss.Style
		isActive := t.Name() == currentPage.Name()
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(string(t.Name())))
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
