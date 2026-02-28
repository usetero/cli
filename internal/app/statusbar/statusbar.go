// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"time"

	"github.com/usetero/cli/internal/app/statusbar/compliance"
	"github.com/usetero/cli/internal/app/statusbar/quality"
	"github.com/usetero/cli/internal/app/statusbar/services"
	"github.com/usetero/cli/internal/app/statusbar/syncstatus"
	"github.com/usetero/cli/internal/app/statusbar/waste"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
)

const diag = "╱"
const workspaceCountTimeout = 2 * time.Second

type workspaceCountMsg struct {
	count int64
	err   error
}

// Tab indices for the drawer.
const (
	TabWaste      = 0
	TabQuality    = 1
	TabCompliance = 2
	TabServices   = 3
	TabSync       = 4
	tabCount      = 5
)

// Tab labels.
var tabLabels = [tabCount]string{"Waste", "Quality", "Compliance", "Services", "Sync"}

// Model renders the app status bar.
type Model struct {
	theme            styles.Theme
	scope            log.Scope
	env              string
	tabs             []drawerTab
	syncStatus       *syncstatus.Model
	servicesStatus   *services.Model
	wasteStatus      *waste.Model
	qualityStatus    *quality.Model
	complianceStatus *compliance.Model
	width            int

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
		theme:            theme,
		scope:            scope,
		env:              env,
		syncStatus:       syncstatus.New(theme, scope, syncer, host),
		servicesStatus:   services.New(theme, scope),
		wasteStatus:      waste.New(theme, scope),
		qualityStatus:    quality.New(theme, scope),
		complianceStatus: compliance.New(theme, scope),
	}
	m.tabs = m.buildTabs()
	return m
}
