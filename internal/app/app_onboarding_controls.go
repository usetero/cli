package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/chat/usecase"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding"
	"github.com/usetero/cli/internal/app/statusbar"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
)

// activateOrg sets the active org, reloads org prefs/storage for the new
// org's isolated data directory, and forwards the message to onboarding.
func (m *Model) activateOrg(orgID domain.OrganizationID, msg tea.Msg) tea.Cmd {
	env := m.cfg.Environment()
	m.scope.Info("setting active org", "org_id", orgID)
	_ = m.userPrefs.SetActiveOrgID(orgID)

	cfg, err := config.LoadOrgPreferences(env, orgID)
	if err != nil {
		m.scope.Error("failed to reload config for org", "error", err)
	} else {
		m.orgPrefs = preferences.NewOrgService(cfg, m.scope)
		m.storage = sqlite.NewStorageService(cfg)
		if m.onboarding != nil {
			m.onboarding.SetOrgPreferences(m.orgPrefs)
		}
	}

	if m.onboarding != nil {
		return m.onboarding.Update(msg)
	}
	return nil
}

// setTheme persists the user's theme preference and shows a toast.
func (m *Model) setTheme(theme preferences.Theme) tea.Cmd {
	if err := m.userPrefs.SetTheme(theme); err != nil {
		m.scope.Error("failed to set theme", "error", err)
		return appmsg.ErrorCmd("Failed to save theme", err, false)
	}
	label := "Auto"
	switch theme {
	case preferences.ThemeAuto:
		// default
	case preferences.ThemeDark:
		label = "Dark"
	case preferences.ThemeLight:
		label = "Light"
	}
	return appmsg.SuccessCmd("Theme set to " + label + ". Restart to apply.")
}

// switchOrganization re-enters onboarding at org selection.
func (m *Model) switchOrganization() tea.Cmd {
	m.scope.Info("switching organization")
	_ = m.userPrefs.SetActiveOrgID("") // clear so onboarding shows the picker
	return m.restartOnboarding()
}

// switchAccount clears account preference (cascades to workspace) and
// re-enters onboarding. The saved org auto-selects, then prompts for account.
func (m *Model) switchAccount() tea.Cmd {
	m.scope.Info("switching account")
	_ = m.orgPrefs.ClearDefaultAccountID()
	return m.restartOnboarding()
}

// restartOnboarding tears down the current session and re-enters onboarding
// at the org selection step. Onboarding auto-selects any saved preferences
// and prompts for anything that was cleared.
func (m *Model) restartOnboarding() tea.Cmd {
	m.shutdown()

	m.db = nil
	m.uploader = nil
	m.chatClient = nil
	m.runtimeDeps = usecase.RuntimeDeps{}
	m.toolRegistry = nil
	m.chat = nil
	m.services = m.services.WithAccountID("") // clear stale account scope

	m.statusBar = statusbar.New(m.theme, m.scope, m.syncer, m.cfg.APIEndpoint, m.cfg.Env)
	m.windowTitle = ""

	m.onboarding = onboarding.New(m.ctx, m.theme, m.services, m.userPrefs, m.orgPrefs, m.authService, m.syncer, m.scope)
	m.state = stateOnboarding
	m.updateLayout()

	return m.onboarding.StartFromOrgSelect()
}
