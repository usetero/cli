// Package onboarding provides the onboarding flow for new users.
package onboarding

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	iauth "github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// Model is the onboarding orchestrator.
type Model struct {
	// Dependencies available from start
	ctx       context.Context
	theme     styles.Theme
	services  graphql.ServiceSet
	userPrefs preferences.UserPreferences
	orgPrefs  preferences.OrgPreferences
	auth      iauth.Auth
	syncer    powersync.Syncer
	scope     log.Scope

	// Accumulated state from step completions
	state bootstrap.State

	// Current step
	gate          Gate
	gateEnteredAt time.Time
	step          Step
	width         int
	height        int

	// Optional test hook for forcing gate build failures in integration tests.
	gateBuildHook func(Gate) error
}

// New creates a new onboarding model.
func New(
	ctx context.Context,
	theme styles.Theme,
	services graphql.ServiceSet,
	userPrefs preferences.UserPreferences,
	orgPrefs preferences.OrgPreferences,
	authService iauth.Auth,
	syncer powersync.Syncer,
	scope log.Scope,
) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	if userPrefs == nil {
		panic("userPrefs is nil")
	}
	if orgPrefs == nil {
		panic("orgPrefs is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}

	scope = scope.Child("onboarding")

	return &Model{
		ctx:       ctx,
		theme:     theme,
		services:  services,
		userPrefs: userPrefs,
		orgPrefs:  orgPrefs,
		auth:      authService,
		syncer:    syncer,
		scope:     scope,
	}
}

// SetOrgPreferences replaces the org preferences used by subsequent onboarding steps.
// Called by the app when switching to a different organization.
func (m *Model) SetOrgPreferences(prefs preferences.OrgPreferences) {
	m.orgPrefs = prefs
}

// Init starts the onboarding flow with preflight resolution.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("onboarding started")
	return m.goToGate(bootstrap.GatePreflight, "init")
}

// StartFromOrgSelect starts the onboarding flow at the organization selection
// step, skipping role selection. Used when switching orgs or
// accounts — the caller clears the relevant preferences so onboarding prompts
// for selection instead of auto-selecting.
func (m *Model) StartFromOrgSelect() tea.Cmd {
	m.scope.Info("onboarding started from org select")
	return m.goToGate(bootstrap.GateOrgSelect, "start_from_org_select")
}

// TestingCurrentGate returns the active onboarding gate.
// For test assertions only.
func (m *Model) TestingCurrentGate() Gate {
	return m.gate
}

// SetTestingGateBuildHook injects a gate build failure hook.
// For tests only.
func (m *Model) SetTestingGateBuildHook(hook func(Gate) error) {
	m.gateBuildHook = hook
}
