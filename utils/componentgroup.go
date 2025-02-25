package utils

import (
	tea "github.com/charmbracelet/bubbletea"
)

// A `tea.Model` that represents an identifiable component in the UI.
type Component interface {
	tea.Model

	ID() string
}

// A collection of `Component` instances that can be managed as a group,
// with the ability to "focus" on a specific component at a time.
//
// Each component is identified by it's `ID()` method.
type ComponentGroup struct {
	focus      int
	components []Component
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

// FocusOn will set the component with the given id as the focused component.
//
// It sends a `utils.FocusMsg{ID: id}` to the focused component, and a `utils.BlurMsg{ID: id}`
// to all other components.
func (c *ComponentGroup) FocusOn(id string) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.components))
	for i, comp := range c.components {
		compID := comp.ID()
		var msg tea.Msg
		if comp.ID() == id {
			msg = FocusMsg{ID: compID}
			c.focus = i
		} else {
			msg = BlurMsg{ID: compID}
		}
		var m tea.Model
		m, cmd := comp.Update(msg)
		c.components[i] = m.(Component)
		cmds[i] = cmd
	}
	return tea.Batch(cmds...)
}

// FocusNext behaves the same as FocusOn but will focus the next sequential component in the group,
// based on the order originally provided to `NewComponentGroup`.
//
// Wraps to the beginning of the group if the last component is focused.
func (c *ComponentGroup) FocusNext() tea.Cmd {
	nextIndex := (c.focus + 1) % len(c.components)
	return c.FocusOn(c.components[nextIndex].ID())
}

// FocusPrevious behaves the same as FocusOn but will focus the previous sequential component in the group,
// based on the order originally provided to `NewComponentGroup`.
//
// Wraps to the end of the group if the first component is focused.
func (c *ComponentGroup) FocusPrevious() tea.Cmd {
	previousIndex := c.focus - 1
	if previousIndex < 0 {
		previousIndex = len(c.components) - 1
	}
	return c.FocusOn(c.components[previousIndex].ID())
}
