package statusbar

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/sqlite"
)

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

type drawerTabAdapter struct {
	label       string
	setDB       func(db sqlite.DB) tea.Cmd
	init        func() tea.Cmd
	update      func(msg tea.Msg) tea.Cmd
	hasData     func() bool
	compactView func() string
	expanded    func(width, height int) string
	interactive bool
	inDetail    func() bool
	closeDetail func()
	handleKey   func(msg tea.KeyPressMsg) tea.Cmd
}

func (m *Model) buildTabs() []drawerTab {
	return []drawerTab{
		drawerTabAdapter{
			label:       tabLabels[TabWaste],
			setDB:       m.wasteStatus.SetDB,
			init:        m.wasteStatus.Init,
			update:      m.wasteStatus.Update,
			hasData:     m.wasteStatus.HasData,
			compactView: m.wasteStatus.CompactView,
			expanded:    m.wasteStatus.ExpandedView,
			interactive: true,
			inDetail:    m.wasteStatus.InDetail,
			closeDetail: m.wasteStatus.CloseDetail,
			handleKey:   m.wasteStatus.HandleKeyPress,
		},
		drawerTabAdapter{
			label:       tabLabels[TabQuality],
			setDB:       m.qualityStatus.SetDB,
			init:        m.qualityStatus.Init,
			update:      m.qualityStatus.Update,
			hasData:     m.qualityStatus.HasData,
			compactView: m.qualityStatus.CompactView,
			expanded:    m.qualityStatus.ExpandedView,
			interactive: true,
			inDetail:    m.qualityStatus.InDetail,
			closeDetail: m.qualityStatus.CloseDetail,
			handleKey:   m.qualityStatus.HandleKeyPress,
		},
		drawerTabAdapter{
			label:       tabLabels[TabCompliance],
			setDB:       m.complianceStatus.SetDB,
			init:        m.complianceStatus.Init,
			update:      m.complianceStatus.Update,
			hasData:     m.complianceStatus.HasData,
			compactView: m.complianceStatus.CompactView,
			expanded:    m.complianceStatus.ExpandedView,
			interactive: true,
			inDetail:    m.complianceStatus.InDetail,
			closeDetail: m.complianceStatus.CloseDetail,
			handleKey:   m.complianceStatus.HandleKeyPress,
		},
		drawerTabAdapter{
			label:       tabLabels[TabServices],
			setDB:       m.servicesStatus.SetDB,
			init:        m.servicesStatus.Init,
			update:      m.servicesStatus.Update,
			hasData:     m.servicesStatus.HasData,
			compactView: m.servicesStatus.CompactView,
			expanded:    m.servicesStatus.ExpandedView,
			interactive: true,
			inDetail:    m.servicesStatus.InDetail,
			closeDetail: m.servicesStatus.CloseDetail,
			handleKey:   m.servicesStatus.HandleKeyPress,
		},
		drawerTabAdapter{
			label:       tabLabels[TabSync],
			setDB:       m.syncStatus.SetDB,
			init:        m.syncStatus.Init,
			update:      m.syncStatus.Update,
			hasData:     m.syncStatus.HasData,
			compactView: m.syncStatus.CompactView,
			expanded:    m.syncStatus.ExpandedView,
			interactive: false,
		},
	}
}

func (t drawerTabAdapter) Label() string { return t.label }

func (t drawerTabAdapter) SetDB(db sqlite.DB) tea.Cmd { return t.setDB(db) }

func (t drawerTabAdapter) Init() tea.Cmd { return t.init() }

func (t drawerTabAdapter) Update(msg tea.Msg) tea.Cmd { return t.update(msg) }

func (t drawerTabAdapter) HasData() bool { return t.hasData() }

func (t drawerTabAdapter) CompactView() string { return t.compactView() }

func (t drawerTabAdapter) ExpandedView(width, height int) string { return t.expanded(width, height) }

func (t drawerTabAdapter) Interactive() bool { return t.interactive }

func (t drawerTabAdapter) InDetail() bool {
	if t.inDetail == nil {
		return false
	}
	return t.inDetail()
}

func (t drawerTabAdapter) CloseDetail() {
	if t.closeDetail != nil {
		t.closeDetail()
	}
}

func (t drawerTabAdapter) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if t.handleKey == nil {
		return nil
	}
	return t.handleKey(msg)
}
