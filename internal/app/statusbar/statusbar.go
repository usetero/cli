// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"github.com/usetero/cli/internal/app/statusbar/services"
	"github.com/usetero/cli/internal/app/statusbar/surfaces"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

const diag = "╱"

// Tab indices for the drawer.
const (
	TabIssues        = 0
	TabChecks        = 1
	TabServices      = 2
	TabLogEvents     = 3
	TabEdgeInstances = 4
	tabCount         = 5
)

// Tab labels.
var tabLabels = [tabCount]string{"Issues", "Checks", "Services", "Log events", "Edge instances"}

// Model renders the app status bar.
type Model struct {
	theme           styles.Theme
	scope           log.Scope
	env             string
	tabs            []drawerTab
	issuesStatus    *surfaces.Model
	checksStatus    *surfaces.Model
	servicesStatus  *services.Model
	logEventsStatus *surfaces.Model
	edgeStatus      *surfaces.Model
	width           int

	// Account context
	org       string
	workspace string

	// Conversation
	title string

	// Context window usage (0-100)
	contextPercent int

	// Drawer state
	drawerOpen bool
	activeTab  int
}

// New creates a new statusbar.
func New(theme styles.Theme, scope log.Scope, host string, env string) *Model {
	_ = host
	scope = scope.Child("statusbar")
	m := &Model{
		theme:           theme,
		scope:           scope,
		env:             env,
		issuesStatus:    surfaces.NewIssues(theme, scope),
		checksStatus:    surfaces.NewChecks(theme, scope),
		servicesStatus:  services.New(theme, scope),
		logEventsStatus: surfaces.NewLogEvents(theme, scope),
		edgeStatus:      surfaces.NewEdgeInstances(theme, scope),
	}
	m.tabs = m.buildTabs()
	return m
}
