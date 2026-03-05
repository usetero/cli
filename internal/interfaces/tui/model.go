package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
)

type state int

const (
	stateOnboarding state = iota
	stateWelcome
)

// Model is the root TUI model for the new architecture.
type Model struct {
	state state
	quit  bool
	scope logging.Scope
}

// New builds the onboarding-first TUI shell.
func New(scope logging.Scope) *Model {
	return &Model{
		state: stateOnboarding,
		scope: scope,
	}
}

// Init starts the program.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("tui initialized")
	return nil
}

// Update handles input and state transitions.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := strings.ToLower(msg.String())
		if s == "ctrl+c" || s == "q" {
			m.scope.Info("quit requested")
			m.quit = true
			return m, tea.Quit
		}
		if s == "enter" && m.state == stateOnboarding {
			m.scope.Info("onboarding completed")
			m.state = stateWelcome
		}
	}
	return m, nil
}

// View renders the current state.
func (m *Model) View() tea.View {
	if m.quit {
		return tea.NewView("")
	}

	switch m.state {
	case stateOnboarding:
		return tea.NewView("Onboarding\n\nPress Enter to complete onboarding.\nPress q to quit.\n")
	case stateWelcome:
		return tea.NewView("Welcome to Taro\n\nPress q to quit.\n")
	default:
		return tea.NewView("")
	}
}
