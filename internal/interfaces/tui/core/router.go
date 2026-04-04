package core

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Router is a small helper for one-active-child composition.
// The parent still owns route selection; Router only owns delegation.
type Router struct {
	active Model
}

// SetActive replaces the current active child.
func (r *Router) SetActive(model Model) {
	r.active = model
}

// Active returns the current active child.
func (r *Router) Active() Model {
	return r.active
}

// Init initializes the active child.
func (r *Router) Init() tea.Cmd {
	if r.active == nil {
		return nil
	}
	return r.active.Init()
}

// Update forwards a message to the active child and keeps its returned model.
func (r *Router) Update(msg tea.Msg) tea.Cmd {
	if r.active == nil {
		return nil
	}
	next, cmd := r.active.Update(msg)
	if typed, ok := next.(Model); ok && typed != nil {
		r.active = typed
	}
	return cmd
}

// View renders the active child.
func (r *Router) View() tea.View {
	if r.active == nil {
		return tea.NewView("")
	}
	return r.active.View()
}

// Page returns the active child's page metadata.
func (r *Router) Page() Page {
	if r.active == nil {
		return Page{}
	}
	provider, ok := r.active.(PageProvider)
	if !ok {
		return Page{}
	}
	return provider.Page()
}

// Commands returns the active child's available shell commands.
func (r *Router) Commands() []Command {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(CommandProvider)
	if !ok {
		return nil
	}
	return provider.Commands()
}

// SetSize forwards layout to the active child only.
func (r *Router) SetSize(width, height int) {
	if r.active == nil {
		return
	}
	r.active.SetSize(width, height)
}

// ShortHelp returns the active child's help bindings.
func (r *Router) ShortHelp() []key.Binding {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(HelpProvider)
	if !ok {
		return nil
	}
	return provider.ShortHelp()
}

// Input returns the active child's input contract.
func (r *Router) Input() *Input {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(InputProvider)
	if !ok {
		return nil
	}
	return provider.Input()
}

// Busy returns the active child's busy state.
func (r *Router) Busy() *Busy {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(BusyProvider)
	if !ok {
		return nil
	}
	return provider.Busy()
}

// Error returns the active child's error state.
func (r *Router) Error() *Error {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(ErrorProvider)
	if !ok {
		return nil
	}
	return provider.Error()
}

// Notice returns the active child's inline notice state.
func (r *Router) Notice() *Notice {
	if r.active == nil {
		return nil
	}
	provider, ok := r.active.(NoticeProvider)
	if !ok {
		return nil
	}
	return provider.Notice()
}
