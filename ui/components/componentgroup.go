package components

import (
	"github.com/alex-laycalvert/ghtui/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	tea.Model

	ID() string
}

type ComponentGroup struct {
	focus      int
	components []Component
}

type ComponentUpdateSizeMsg struct {
	Width  int
	Height int
}

func NewComponentGroup(components ...Component) ComponentGroup {
	group := ComponentGroup{components: components}
	return group
}

func (c ComponentGroup) GetComponent(id string) Component {
	for _, comp := range c.components {
		if comp.ID() == id {
			return comp
		}
	}
	return nil
}

func (c ComponentGroup) GetComponents() []Component {
	return c.components
}

func (c ComponentGroup) Init() tea.Cmd {
	initCmds := make([]tea.Cmd, len(c.components))
	for i, comp := range c.components {
		initCmds[i] = comp.Init()
	}
	return tea.Batch(initCmds...)
}

func (c *ComponentGroup) Update(id string, msg tea.Msg) tea.Cmd {
	for i, comp := range c.components {
		if comp.ID() == id {
			m, cmd := comp.Update(msg)
			c.components[i] = m.(Component)
			return cmd
		}
	}
	return nil
}

func (c *ComponentGroup) UpdateAll(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.components))
	for i, comp := range c.components {
		m, cmd := comp.Update(msg)
		c.components[i] = m.(Component)
		cmds[i] = cmd
	}
	return tea.Batch(cmds...)
}

func (c *ComponentGroup) UpdateFocused(msg tea.Msg) tea.Cmd {
	if c.focus == -1 {
		return nil
	}

	m, cmd := c.components[c.focus].Update(msg)
	c.components[c.focus] = m.(Component)
	return cmd
}

func (c ComponentGroup) GetFocusedComponent() Component {
	if c.focus == -1 {
		return nil
	}
	return c.components[c.focus]
}

func (c ComponentGroup) GetFocusedComponentName() string {
	focusedComponent := c.GetFocusedComponent()
	if focusedComponent == nil {
		return ""
	}
	return focusedComponent.ID()
}

func (c ComponentGroup) IsFocused(id string) bool {
	return c.GetFocusedComponentName() == id
}

func (c *ComponentGroup) FocusOn(id string) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.components))
	for i, comp := range c.components {
		var msg tea.Msg
		if comp.ID() == id {
			msg = utils.FocusMsg{ID: id}
			c.focus = i
		} else {
			msg = utils.BlurMsg{ID: id}
		}
		var m tea.Model
		m, cmd := comp.Update(msg)
		c.components[i] = m.(Component)
		cmds[i] = cmd
	}
	return tea.Batch(cmds...)
}

func (c *ComponentGroup) FocusNext() tea.Cmd {
	c.focus = (c.focus + 1) % len(c.components)
	var m tea.Model
	m, cmd := c.components[c.focus].Update(
		utils.FocusMsg{ID: c.components[c.focus].ID()},
	)
	c.components[c.focus] = m.(Component)
	return cmd
}

func (c *ComponentGroup) FocusPrevious() tea.Cmd {
	c.focus = c.focus - 1
	if c.focus < 0 {
		c.focus = len(c.components) - 1
	}
	var m tea.Model
	m, cmd := c.components[c.focus].Update(
		utils.BlurMsg{ID: c.components[c.focus].ID()},
	)
	c.components[c.focus] = m.(Component)
	return cmd
}
