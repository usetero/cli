package statusbar

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/sqlite"
)

// TabModel is the base contract every status bar tab model must satisfy.
type TabModel interface {
	SetDB(db sqlite.DB) tea.Cmd
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
	SetDB(db sqlite.DB) tea.Cmd
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

type tab struct {
	label string
	model TabModel
}

func newTab(label string, model TabModel) drawerTab {
	return tab{label: label, model: model}
}

func (t tab) Label() string { return t.label }

func (t tab) SetDB(db sqlite.DB) tea.Cmd { return t.model.SetDB(db) }

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
		newTab(tabLabels[TabWaste], m.wasteStatus),
		newTab(tabLabels[TabQuality], m.qualityStatus),
		newTab(tabLabels[TabCompliance], m.complianceStatus),
		newTab(tabLabels[TabServices], m.servicesStatus),
		newTab(tabLabels[TabSync], m.syncStatus),
	}
}
