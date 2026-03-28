package core

import tea "charm.land/bubbletea/v2"

// Model is the common contract for TUI models.
//
// Rules:
// - Update handles local UI state and returns typed messages.
// - SetSize applies layout changes from parent WindowSize events.
type Model interface {
	tea.Model
	SetSize(width, height int)
}
