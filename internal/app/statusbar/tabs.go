package statusbar

import (
	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
)

// TabModel is the base contract every status bar tab model must satisfy.
// Tabs read from control-plane GraphQL services, not a local database.
type TabModel interface {
	SetServices(services graphql.ServiceSet) tea.Cmd
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	HasData() bool
	CompactView() string
	ExpandedView(width, height int) string
}

// InteractiveTabModel extends TabModel with drawer interaction behavior.
type InteractiveTabModel interface {
	TabModel
	InDetail() bool
	CloseDetail()
	HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd
}

type drawerTab interface {
	Label() string
	SetServices(services graphql.ServiceSet) tea.Cmd
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	HasData() bool
	CompactView() string
	ExpandedView(width, height int) string
	Interactive() bool
	InDetail() bool
	CloseDetail()
	HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd
}

type groupedDrawerTab interface {
	GroupLabel() string
}

type tab struct {
	group string
	label string
	model TabModel
}

func newTab(group, label string, model TabModel) drawerTab {
	return tab{group: group, label: label, model: model}
}

func (t tab) Label() string { return t.label }

func (t tab) GroupLabel() string { return t.group }

func (t tab) SetServices(services graphql.ServiceSet) tea.Cmd { return t.model.SetServices(services) }

func (t tab) Init() tea.Cmd { return t.model.Init() }

func (t tab) Update(msg tea.Msg) tea.Cmd { return t.model.Update(msg) }

func (t tab) HasData() bool { return t.model.HasData() }

func (t tab) CompactView() string { return t.model.CompactView() }

func (t tab) ExpandedView(width, height int) string { return t.model.ExpandedView(width, height) }

func (t tab) Interactive() bool {
	_, ok := t.model.(InteractiveTabModel)
	return ok
}

func (t tab) InDetail() bool {
	interactive, ok := t.model.(InteractiveTabModel)
	if !ok {
		return false
	}
	return interactive.InDetail()
}

func (t tab) CloseDetail() {
	if interactive, ok := t.model.(InteractiveTabModel); ok {
		interactive.CloseDetail()
	}
}

func (t tab) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	interactive, ok := t.model.(InteractiveTabModel)
	if !ok {
		return nil
	}
	return interactive.HandleKeyPress(msg)
}

func (m *Model) buildTabs() []drawerTab {
	return []drawerTab{
		newTab("Control Plane", tabLabels[TabIssues], m.issuesStatus),
		newTab("Control Plane", tabLabels[TabChecks], m.checksStatus),
		newTab("Data Plane", tabLabels[TabServices], m.servicesStatus),
		newTab("Data Plane", tabLabels[TabLogEvents], m.logEventsStatus),
		newTab("Data Plane", tabLabels[TabEdgeInstances], m.edgeStatus),
	}
}
