package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	tea.Model

	Name() string
}

type ComponentGroup struct {
	focus      int
	components []Component
}

func NewComponentGroup(components []Component) ComponentGroup {
	group := ComponentGroup{components: components}

	return group
}

func (c ComponentGroup) GetComponent(name string) Component {
	for _, comp := range c.components {
		if comp.Name() == name {
			return comp
		}
	}
	return nil
}

func (c ComponentGroup) Init() tea.Cmd {
	initCmds := make([]tea.Cmd, len(c.components))
	for i, comp := range c.components {
		initCmds[i] = comp.Init()
	}
	return tea.Batch(initCmds...)
}

func (c *ComponentGroup) Update(name string, msg tea.Msg) tea.Cmd {
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

func (c *ComponentGroup) FocusOn(name string) {
	for i, comp := range c.components {
		if comp.Name() == name {
			c.focus = i
			break
		}
	}
}

func NameComponent[T tea.Model](name string, component T) NamedComponent[T] {
	return NamedComponent[T]{
		name,
		component,
	}
}

type NamedComponent[T tea.Model] struct {
	name      string
	Component T
}

func (w NamedComponent[T]) Name() string {
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
