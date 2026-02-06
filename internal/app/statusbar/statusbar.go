// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/catalogstatus"
	"github.com/usetero/cli/internal/app/statusbar/syncstatus"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const diag = "╱"

// Tab indices for the drawer.
const (
	TabSync    = 0
	TabCatalog = 1
	TabChat    = 2
	tabCount   = 3
)

// Tab labels.
var tabLabels = [tabCount]string{"Sync", "Catalog", "Chat"}

// Model renders the app status bar.
type Model struct {
	theme         *styles.Theme
	syncStatus    *syncstatus.Model
	catalogStatus *catalogstatus.Model
	width         int

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
func New(theme *styles.Theme, syncer powersync.Syncer, host string) *Model {
	return &Model{
		theme:         theme,
		syncStatus:    syncstatus.New(theme, syncer, host),
		catalogStatus: catalogstatus.New(theme),
	}
}

// SetDB sets the database for catalog status polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	return m.catalogStatus.SetDB(db)
}

// Init initializes child models.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.syncStatus.Init(),
		m.catalogStatus.Init(),
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	return tea.Batch(
		m.syncStatus.Update(msg),
		m.catalogStatus.Update(msg),
	)
}

// SetWidth sets the statusbar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetOrg sets the organization name.
func (m *Model) SetOrg(org string) {
	m.org = org
}

// SetWorkspace sets the workspace name.
func (m *Model) SetWorkspace(workspace string) {
	m.workspace = workspace
}

// SetTitle sets the conversation title.
func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetContextPercent sets the context window usage percentage.
func (m *Model) SetContextPercent(percent int) {
	m.contextPercent = percent
}

// ToggleDrawer toggles the drawer open/closed.
func (m *Model) ToggleDrawer() {
	m.drawerOpen = !m.drawerOpen
}

// CloseDrawer closes the drawer.
func (m *Model) CloseDrawer() {
	m.drawerOpen = false
}

// NextTab cycles to the next drawer tab.
func (m *Model) NextTab() {
	m.activeTab = (m.activeTab + 1) % tabCount
}

// IsDrawerOpen returns whether the drawer is open.
func (m *Model) IsDrawerOpen() bool {
	return m.drawerOpen
}

// Height returns the height of the statusbar.
func (m *Model) Height() int {
	return 1
}

// View renders the statusbar.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	colors := m.theme.Colors
	diagStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)

	sep := diagStyle.Render(" " + diag + " ")

	// Build left-aligned segments
	var segments []string

	// 1. Brand + sync status (always shown)
	brandConn := m.renderBrand()
	segments = append(segments, brandConn)

	// 2. Catalog pulse (phase-aware: services, policies, discovery progress)
	catalogView := m.catalogStatus.CompactView()
	if catalogView != "" {
		segments = append(segments, catalogView)
	}

	// 3. Org / workspace (only shown if set)
	if m.org != "" {
		orgWs := m.renderOrgWorkspace()
		segments = append(segments, orgWs)
	}

	// Calculate what fits
	baseContent := strings.Join(segments, sep)
	baseWidth := lipgloss.Width(baseContent)

	// 4. Title (if fits)
	if m.title != "" {
		titleSeg := m.renderTitle(m.width - baseWidth - 20) // leave room for context
		if titleSeg != "" {
			testContent := baseContent + sep + titleSeg
			if lipgloss.Width(testContent) < m.width-15 {
				segments = append(segments, titleSeg)
			}
		}
	}

	// 5. Context window usage (if fits)
	if m.contextPercent > 0 {
		pctSeg := m.renderContextPercent()
		testContent := strings.Join(segments, sep) + sep + pctSeg
		if lipgloss.Width(testContent) < m.width-3 {
			segments = append(segments, pctSeg)
		}
	}

	// Build right-aligned segment: ctrl+d hint
	rightSeg := m.renderDrawerHint()
	rightWidth := lipgloss.Width(rightSeg)

	// Join left segments
	content := strings.Join(segments, sep)
	contentWidth := lipgloss.Width(content)

	// Fill space between left content and right hint with diagonals
	leftDiags := diagStyle.Render(diag + diag)
	middlePadding := m.width - contentWidth - rightWidth - 7 // 2 left diags + 2 right diags + 3 spaces
	if middlePadding < 3 {
		middlePadding = 3
	}
	middleDiags := diagStyle.Render(strings.Repeat(diag, middlePadding))
	rightDiags := diagStyle.Render(diag + diag)

	return leftDiags + " " + content + " " + middleDiags + " " + rightSeg + rightDiags
}

// DrawerView renders the drawer overlay.
func (m *Model) DrawerView(width, height int) string {
	colors := m.theme.Colors

	// Tab bar
	tabBar := m.renderTabBar(width - 4) // account for border + padding

	// Active tab content
	var content string
	switch m.activeTab {
	case TabSync:
		content = m.syncStatus.ExpandedView()
	case TabCatalog:
		content = m.catalogStatus.ExpandedView()
	case TabChat:
		content = m.renderContextPercent()
	}

	if content == "" {
		content = lipgloss.NewStyle().Foreground(colors.Page.TextSubtle).Render("No data")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content)

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Accent).
		Padding(0, 1)

	return style.Render(inner)
}

// renderTabBar renders the tab selector for the drawer.
func (m *Model) renderTabBar(width int) string {
	colors := m.theme.Colors
	activeStyle := lipgloss.NewStyle().Foreground(colors.Accent).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	sepStyle := lipgloss.NewStyle().Foreground(colors.Page.TextSubtle)

	var tabs []string
	for i, label := range tabLabels {
		if i == m.activeTab {
			tabs = append(tabs, activeStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveStyle.Render(label))
		}
	}

	return strings.Join(tabs, sepStyle.Render("  "))
}

// renderDrawerHint renders the "ctrl+d open/close" hint.
func (m *Model) renderDrawerHint() string {
	colors := m.theme.Colors
	keyStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	tipStyle := lipgloss.NewStyle().Foreground(colors.Page.TextSubtle)

	tip := " open "
	if m.drawerOpen {
		tip = " close"
	}

	return keyStyle.Render("ctrl+d") + tipStyle.Render(tip)
}

// renderBrand renders "TERO" with optional sync status.
func (m *Model) renderBrand() string {
	colors := m.theme.Colors

	brand := styles.ApplyBoldForegroundGrad("TERO", colors.Brand.GradientStart, colors.Brand.GradientEnd)

	syncView := m.syncStatus.CompactView()
	if syncView == "" {
		return brand
	}

	return brand + " " + syncView
}

// renderOrgWorkspace renders "org / workspace".
func (m *Model) renderOrgWorkspace() string {
	colors := m.theme.Colors
	style := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	if m.workspace != "" {
		return style.Render(m.org + " / " + m.workspace)
	}
	return style.Render(m.org)
}

// renderTitle renders the conversation title, truncated to maxWidth.
func (m *Model) renderTitle(maxWidth int) string {
	if maxWidth < 10 {
		return ""
	}

	colors := m.theme.Colors
	style := lipgloss.NewStyle().Foreground(colors.Page.Text)

	title := m.title
	if len(title) > maxWidth-2 {
		title = title[:maxWidth-5] + "..."
	}

	return style.Render("\"" + title + "\"")
}

// renderContextPercent renders "ctx: N%" with color based on usage level.
func (m *Model) renderContextPercent() string {
	colors := m.theme.Colors

	var fg color.Color
	switch {
	case m.contextPercent >= 90:
		fg = colors.Error.Fg
	case m.contextPercent >= 75:
		fg = colors.Warning.Fg
	default:
		fg = colors.Page.TextMuted
	}

	style := lipgloss.NewStyle().Foreground(fg)
	return style.Render(fmt.Sprintf("ctx: %d%%", m.contextPercent))
}
