package core

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Children is a small helper for applying common operations across active child models.
type Children []Model

// Init initializes all child models in order.
func (c Children) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(c))
	for _, model := range c {
		if model == nil {
			continue
		}
		if cmd := model.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// Update delivers a message to all child models in order.
func (c Children) Update(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(c))
	for _, model := range c {
		if model == nil {
			continue
		}
		if _, cmd := model.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// ShortHelp aggregates short-help bindings from all child models.
func (c Children) ShortHelp() []key.Binding {
	var bindings []key.Binding
	for _, model := range c {
		if model == nil {
			continue
		}
		provider, ok := model.(HelpProvider)
		if !ok {
			continue
		}
		bindings = append(bindings, provider.ShortHelp()...)
	}
	return bindings
}
