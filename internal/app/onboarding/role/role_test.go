package role

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
)

func TestInitUsesSavedRolePreference(t *testing.T) {
	t.Parallel()

	prefs := preferencestest.NewMockUserPreferences()
	prefs.GetRoleFunc = func() string { return bootstrap.RoleEngineer }
	m := New(styles.NewTheme(true), prefs, logtest.NewScope(t))

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected init cmd")
	}

	updateCmd := m.Update(cmd())
	if updateCmd == nil {
		t.Fatal("expected saved role to emit RoleSelected")
	}
	if selected, ok := updateCmd().(bootstrap.RoleSelected); !ok || selected.Role != bootstrap.RoleEngineer {
		t.Fatalf("expected RoleSelected(engineer), got %#v", selected)
	}
}

func TestEnterSelectsRoleAndPersistsPreference(t *testing.T) {
	t.Parallel()

	var saved string
	prefs := preferencestest.NewMockUserPreferences()
	prefs.SetRoleFunc = func(role string) error {
		saved = role
		return nil
	}
	m := New(styles.NewTheme(true), prefs, logtest.NewScope(t))

	// Move to second option (engineer), then select.
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected role selection command")
	}

	msg := cmd()
	selected, ok := msg.(bootstrap.RoleSelected)
	if !ok {
		t.Fatalf("expected RoleSelected message, got %T", msg)
	}
	if selected.Role != bootstrap.RoleEngineer {
		t.Fatalf("expected engineer role, got %q", selected.Role)
	}
	if saved != bootstrap.RoleEngineer {
		t.Fatalf("expected prefs to persist engineer role, got %q", saved)
	}
}
