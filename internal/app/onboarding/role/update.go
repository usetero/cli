package role

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case savedRoleLoadedMsg:
		if msg.role == bootstrap.RolePlatform || msg.role == bootstrap.RoleEngineer {
			m.scope.Debug("using saved role preference", "role", msg.role)
			return func() tea.Msg {
				return bootstrap.RoleSelected{Role: msg.role}
			}
		}
		if msg.role == bootstrap.RoleEngineer {
			m.selected = 1
		}
		return nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, upKey):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, downKey):
			if m.selected < 1 {
				m.selected++
			}
		case key.Matches(msg, selectKey):
			role := bootstrap.RolePlatform
			if m.selected == 1 {
				role = bootstrap.RoleEngineer
			}
			_ = m.prefs.SetRole(role)
			m.scope.Info("role selected", "role", role)
			return func() tea.Msg {
				return bootstrap.RoleSelected{Role: role}
			}
		}
	}
	return nil
}
