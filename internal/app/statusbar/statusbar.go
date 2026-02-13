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

	onboardingmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/app/statusbar/services"
	"github.com/usetero/cli/internal/app/statusbar/syncstatus"
	"github.com/usetero/cli/internal/app/statusbar/waste"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/keymap"
)

const diag = "╱"

// Tab indices for the drawer.
const (
	TabWaste    = 0
	TabServices = 1
	TabSync     = 2
	tabCount    = 3
)

// Tab labels.
var tabLabels = [tabCount]string{"Waste", "Services", "Sync"}

// Model renders the app status bar.
type Model struct {
	theme          styles.Theme
	env            string
	syncStatus     *syncstatus.Model
	servicesStatus *services.Model
	wasteStatus    *waste.Model
	width          int

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
func New(theme styles.Theme, syncer powersync.Syncer, host string, env string) *Model {
	return &Model{
		theme:          theme,
		env:            env,
		syncStatus:     syncstatus.New(theme, syncer, host),
		servicesStatus: services.New(theme),
		wasteStatus:    waste.New(theme),
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
		m.servicesStatus.SetDB(db),
		m.wasteStatus.SetDB(db),
	)
}

// Init initializes child models.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.syncStatus.Init(),
		m.servicesStatus.Init(),
		m.wasteStatus.Init(),
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case onboardingmsg.OrgSelected:
		m.org = msg.Org.Name
	case onboardingmsg.OrgCreated:
		m.org = msg.Org.Name
	case onboardingmsg.WorkspaceSelected:
		m.workspace = msg.Workspace.Name
	}

	return tea.Batch(
		m.syncStatus.Update(msg),
		m.servicesStatus.Update(msg),
		m.wasteStatus.Update(msg),
	)
}

// SetWidth sets the statusbar width.
func (m *Model) SetWidth(width int) {
	m.width = width
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
// Opening is suppressed until at least one tab has data.
func (m *Model) ToggleDrawer() {
	if m.drawerOpen {
		m.drawerOpen = false
		return
	}
	if m.servicesStatus.HasData() || m.wasteStatus.HasData() {
		m.drawerOpen = true
	}
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

	// 2. Services health (dot + service count + discovery)
	servicesView := m.servicesStatus.CompactView()
	if servicesView != "" {
		segments = append(segments, servicesView)
	}

	// 3. Waste status (pending count, estimated/observed savings)
	wasteView := m.wasteStatus.CompactView()
	if wasteView != "" {
		segments = append(segments, wasteView)
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

	if rightSeg == "" {
		// No right segment — fill remaining space with diagonals
		trailing := m.width - contentWidth - 3 // 2 left diags + 1 space
		if trailing < 1 {
			trailing = 1
		}
		return leftDiags + " " + content + " " + diagStyle.Render(strings.Repeat(diag, trailing))
	}

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
	case TabServices:
		content = m.servicesStatus.ExpandedView(contentWidth, contentHeight)
	case TabWaste:
		content = m.wasteStatus.ExpandedView(contentWidth, contentHeight)
	}

	if content == "" {
		content = lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg).Render("Waiting for data...")
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
	// Hide hint until data is loaded and the drawer can open.
	if !m.drawerOpen && !m.servicesStatus.HasData() && !m.wasteStatus.HasData() {
		return ""
	}

	colors := m.theme
	keyStyle := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	tipStyle := lipgloss.NewStyle().Foreground(colors.TextSubtle).Background(colors.Bg)

	tip := " open "
	if m.drawerOpen {
		tip = " close"
	}

	return keyStyle.Render("ctrl+d") + tipStyle.Render(tip)
}

// renderBrand renders "TERO [env] ● Org" (sync dot + org context as one unit).
func (m *Model) renderBrand() string {
	colors := m.theme

	brand := styles.ApplyBoldForegroundGrad("TERO", colors.GradientStart, colors.GradientEnd)

	if m.env != "" && m.env != "prd" {
		envStyle := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		brand += " " + envStyle.Render(strings.ToUpper(m.env))
	}

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
		fg = colors.Error
	case m.contextPercent >= 75:
		fg = colors.Warning
	default:
		fg = colors.TextMuted
	}

	style := lipgloss.NewStyle().Foreground(fg).Background(colors.Bg)
	return style.Render(fmt.Sprintf("ctx: %d%%", m.contextPercent))
}
