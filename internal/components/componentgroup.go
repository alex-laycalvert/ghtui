package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ComponentName string

type Component interface {
	tea.Model

	Name() ComponentName
}

type ComponentGroup struct {
	focus      int
	components []Component
}

func NewComponentGroup(components ...Component) ComponentGroup {
	group := ComponentGroup{components: components}
	return group
}

func (c ComponentGroup) GetComponent(name ComponentName) Component {
	for _, comp := range c.components {
		if comp.Name() == name {
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

func (c *ComponentGroup) Update(name ComponentName, msg tea.Msg) tea.Cmd {
	for i, comp := range c.components {
		if comp.Name() == name {
			m, cmd := comp.Update(msg)
			c.components[i] = m.(Component)
			return cmd
		}
	}
	return nil
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

func (c *ComponentGroup) FocusOn(name ComponentName) {
	for i, comp := range c.components {
		if comp.Name() == name {
			c.focus = i
			break
		}
	}
}

func (c *ComponentGroup) FocusNext() {
	c.focus = (c.focus + 1) % len(c.components)
}

func (c *ComponentGroup) FocusPrevious() {
	c.focus = c.focus - 1
	if c.focus < 0 {
		c.focus = len(c.components) - 1
	}
}

func NameComponent[T tea.Model](name ComponentName, component T) NamedComponent[T] {
	return NamedComponent[T]{
		name,
		component,
	}
}

type NamedComponent[T tea.Model] struct {
	name      ComponentName
	Component T
}

func (w NamedComponent[T]) Name() ComponentName {
	return w.name
}

func (w NamedComponent[T]) Init() tea.Cmd {
	return w.Component.Init()
}

func (w NamedComponent[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	c, cmd := w.Component.Update(msg)
	w.Component = c.(T)
	return w, cmd
}

func (w NamedComponent[T]) View() string {
	return w.Component.View()
}
