package role

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case savedRoleMsg:
		if msg.role == msgs.RolePlatform || msg.role == msgs.RoleEngineer {
			m.scope.Debug("using saved role preference", "role", msg.role)
			return func() tea.Msg {
				return msgs.RoleSelected{Role: msg.role}
			}
		}
		if msg.role == msgs.RoleEngineer {
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
			role := msgs.RolePlatform
			if m.selected == 1 {
				role = msgs.RoleEngineer
			}
			_ = m.prefs.SetRole(role)
			m.scope.Info("role selected", "role", role)
			return func() tea.Msg {
				return msgs.RoleSelected{Role: role}
			}
		}
	}
	return nil
}
