// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"time"

	"github.com/usetero/cli/internal/app/statusbar/services"
	"github.com/usetero/cli/internal/app/statusbar/surfaces"
	"github.com/usetero/cli/internal/app/statusbar/syncstatus"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
)

const diag = "╱"
const workspaceCountTimeout = 2 * time.Second

type workspaceCountLoadedMsg struct {
	count int64
	err   error
}

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
	syncStatus      *syncstatus.Model
	issuesStatus    *surfaces.Model
	checksStatus    *surfaces.Model
	servicesStatus  *services.Model
	logEventsStatus *surfaces.Model
	edgeStatus      *surfaces.Model
	width           int

	// Account context
	org            string
	workspace      string
	workspaceCount int64

	// Conversation
	title string

	// Context window usage (0-100)
	contextPercent int

	// Drawer state
	drawerOpen bool
	activeTab  int
}

// New creates a new statusbar.
func New(theme styles.Theme, scope log.Scope, syncer powersync.Syncer, host string, env string) *Model {
	scope = scope.Child("statusbar")
	m := &Model{
		theme:           theme,
		scope:           scope,
		env:             env,
		syncStatus:      syncstatus.New(theme, scope, syncer, host),
		issuesStatus:    surfaces.NewIssues(theme, scope),
		checksStatus:    surfaces.NewChecks(theme, scope),
		servicesStatus:  services.New(theme, scope),
		logEventsStatus: surfaces.NewLogEvents(theme, scope),
		edgeStatus:      surfaces.NewEdgeInstances(theme, scope),
	}
	m.tabs = m.buildTabs()
	return m
}
