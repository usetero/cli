// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/catalogstatus"
	"github.com/usetero/cli/internal/app/statusbar/policystatus"
	"github.com/usetero/cli/internal/app/statusbar/syncstatus"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/keymap"
)

const diag = "╱"

// Tab indices for the drawer.
const (
	TabPolicy  = 0
	TabCatalog = 1
	TabSync    = 2
	tabCount   = 3
)

// Tab labels.
var tabLabels = [tabCount]string{"Policy", "Catalog", "Sync"}

// Model renders the app status bar.
type Model struct {
	theme         styles.Theme
	syncStatus    *syncstatus.Model
	catalogStatus *catalogstatus.Model
	policyStatus  *policystatus.Model
	width         int

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
func New(theme styles.Theme, syncer powersync.Syncer, host string) *Model {
	return &Model{
		theme:         theme,
		syncStatus:    syncstatus.New(theme, syncer, host),
		catalogStatus: catalogstatus.New(theme),
		policyStatus:  policystatus.New(theme),
	}
}

// SetDB sets the database for status polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM workspaces")
	_ = row.Scan(&m.workspaceCount)

	return tea.Batch(
		m.syncStatus.SetDB(db),
		m.catalogStatus.SetDB(db),
		m.policyStatus.SetDB(db),
	)
}

// Init initializes child models.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.syncStatus.Init(),
		m.catalogStatus.Init(),
		m.policyStatus.Init(),
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	return tea.Batch(
		m.syncStatus.Update(msg),
		m.catalogStatus.Update(msg),
		m.policyStatus.Update(msg),
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

// ShortHelp returns keybindings shown in the keybar when the drawer is open.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keymap.NextTab, keymap.CloseDrawer}
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

	colors := m.theme
	diagStyle := lipgloss.NewStyle().Foreground(colors.GradientEnd).Background(colors.Bg)

	sep := diagStyle.Render(" " + diag + " ")

	// Build left-aligned segments
	var segments []string

	// 1. Brand + sync dot + org context (always shown)
	segments = append(segments, m.renderBrand())

	// 2. Catalog health (dot + service count or discovery phase)
	catalogView := m.catalogStatus.CompactView()
	if catalogView != "" {
		segments = append(segments, catalogView)
	}

	// 3. Policy status (pending count, estimated/observed savings)
	policyView := m.policyStatus.CompactView()
	if policyView != "" {
		segments = append(segments, policyView)
	}

	// Calculate what fits
	baseContent := strings.Join(segments, sep)
	baseWidth := lipgloss.Width(baseContent)

	// 5. Title (if fits)
	if m.title != "" {
		titleSeg := m.renderTitle(m.width - baseWidth - 20) // leave room for context
		if titleSeg != "" {
			testContent := baseContent + sep + titleSeg
			if lipgloss.Width(testContent) < m.width-15 {
				segments = append(segments, titleSeg)
			}
		}
	}

	// 6. Context window usage (only shown when high)
	if m.contextPercent >= 75 {
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
	colors := m.theme

	// Tab bar
	tabBar := m.renderTabBar(width - 4) // account for border + padding

	// Active tab content
	contentWidth := width - 4   // border (2) + padding (2)
	contentHeight := height - 4 // border (2) + tab bar (1) + gap (1)
	var content string
	switch m.activeTab {
	case TabSync:
		content = m.syncStatus.ExpandedView(contentWidth, contentHeight)
	case TabCatalog:
		content = m.catalogStatus.ExpandedView(contentWidth, contentHeight)
	case TabPolicy:
		content = m.policyStatus.ExpandedView(contentWidth, contentHeight)
	}

	if content == "" {
		content = lipgloss.NewStyle().Foreground(colors.TextSubtle).Background(colors.Bg).Render("No data")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content)

	style := lipgloss.NewStyle().
		Width(width).
		MaxHeight(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.GradientEnd).
		Padding(0, 1)

	return style.Render(inner)
}

// renderTabBar renders the tab selector for the drawer.
func (m *Model) renderTabBar(width int) string {
	colors := m.theme
	activeStyle := lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sepStyle := lipgloss.NewStyle().Foreground(colors.TextSubtle).Background(colors.Bg)

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
	colors := m.theme
	keyStyle := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	tipStyle := lipgloss.NewStyle().Foreground(colors.TextSubtle).Background(colors.Bg)

	tip := " open "
	if m.drawerOpen {
		tip = " close"
	}

	return keyStyle.Render("ctrl+d") + tipStyle.Render(tip)
}

// renderBrand renders "TERO ● Org" (sync dot + org context as one unit).
func (m *Model) renderBrand() string {
	colors := m.theme

	brand := styles.ApplyBoldForegroundGrad("TERO", colors.GradientStart, colors.GradientEnd)

	syncView := m.syncStatus.CompactView()
	if syncView != "" {
		brand += " " + syncView
	}

	if m.org != "" {
		brand += " " + m.renderOrgWorkspace()
	}

	return brand
}

// renderOrgWorkspace renders org context. Includes workspace when multiple exist.
func (m *Model) renderOrgWorkspace() string {
	colors := m.theme
	style := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	if m.workspace != "" && m.workspaceCount > 1 {
		return style.Render(m.org + " / " + m.workspace)
	}
	return style.Render(m.org)
}

// renderTitle renders the conversation title, truncated to maxWidth.
func (m *Model) renderTitle(maxWidth int) string {
	if maxWidth < 10 {
		return ""
	}

	colors := m.theme
	style := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)

	title := m.title
	if len(title) > maxWidth-2 {
		title = title[:maxWidth-5] + "..."
	}

	return style.Render("\"" + title + "\"")
}

// renderContextPercent renders "ctx: N%" with color based on usage level.
func (m *Model) renderContextPercent() string {
	colors := m.theme

	var fg color.Color
	switch {
	case m.contextPercent >= 90:
		fg = colors.ErrorBg
	case m.contextPercent >= 75:
		fg = colors.WarningBg
	default:
		fg = colors.TextMuted
	}

	style := lipgloss.NewStyle().Foreground(fg).Background(colors.Bg)
	return style.Render(fmt.Sprintf("ctx: %d%%", m.contextPercent))
}
