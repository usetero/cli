package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/auth/authtest"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
)

func TestRenderContentOverlayPrecedencePaletteOverQuitDialog(t *testing.T) {
	t.Parallel()

	m := newViewTestModel(t)
	m.quitDlg = newQuitDialog(m.theme)
	cmd := m.openPalette()
	if cmd != nil {
		_ = cmd()
	}

	got := m.renderContent()
	if !strings.Contains(got, "Commands") {
		t.Fatalf("expected palette to be rendered, got: %q", got)
	}
	if strings.Contains(got, "Are you sure you want to quit?") {
		t.Fatalf("expected quit dialog to be hidden behind palette")
	}
}

func TestRenderContentShowsQuitDialogWhenPaletteClosed(t *testing.T) {
	t.Parallel()

	m := newViewTestModel(t)
	m.quitDlg = newQuitDialog(m.theme)

	got := m.renderContent()
	if !strings.Contains(got, "Are you sure you want to quit?") {
		t.Fatalf("expected quit dialog content, got: %q", got)
	}
}

func TestViewSuppressesCursorWhenQuitDialogOpen(t *testing.T) {
	t.Parallel()

	m := newViewTestModel(t)
	cmd := m.openPalette()
	if cmd != nil {
		_ = cmd()
	}

	viewWithPalette := m.View()
	if viewWithPalette.Cursor == nil {
		t.Fatalf("expected palette view to expose cursor")
	}

	m.quitDlg = newQuitDialog(m.theme)
	viewWithQuit := m.View()
	if viewWithQuit.Cursor != nil {
		t.Fatalf("expected cursor to be suppressed when quit dialog is open")
	}
}

func newViewTestModel(t *testing.T) *Model {
	t.Helper()

	scope := logtest.NewScope(t)
	client := apitest.NewMockClient()
	services := graphql.NewServiceSetFromClient(client, scope)
	userPrefs := preferencestest.NewMockUserPreferences()
	orgPrefs := preferencestest.NewMockOrgPreferences()
	authSvc := &authtest.MockAuth{}
	syncer := powersynctest.NewMockSyncer()

	m := New(
		context.Background(),
		&config.CLIConfig{
			Env:       "dev",
			APIOrigin: "https://api.example.com",
		},
		styles.NewTheme(true),
		"dev",
		services,
		authSvc,
		userPrefs,
		orgPrefs,
		testStorage{dbPath: filepath.Join(t.TempDir(), "view-test.sqlite")},
		syncer,
		scope,
	)
	m.width = 120
	m.height = 40
	m.updateLayout()
	return m
}
