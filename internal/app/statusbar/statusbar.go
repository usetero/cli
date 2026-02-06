// Package statusbar renders the persistent top bar for the Tero app.
package statusbar

import (
	"fmt"
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

	// Context
	contextCount int

	// Context window usage (0-100)
	contextPercent int
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

// SetContextCount sets the number of entities in context.
func (m *Model) SetContextCount(count int) {
	m.contextCount = count
}

// SetContextPercent sets the context window usage percentage.
func (m *Model) SetContextPercent(percent int) {
	m.contextPercent = percent
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
	sepStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	sep := sepStyle.Render(" │ ")

	// Build segments from left to right
	var segments []string

	// 1. Brand + sync status (always shown)
	brandConn := m.renderBrand()
	segments = append(segments, brandConn)

	// 2. Catalog pulse (phase-aware: services, policies, discovery progress)
	catalogView := m.catalogStatus.View()
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

	// 5. Context count (if fits)
	if m.contextCount > 0 {
		ctxSeg := mutedStyle.Render(fmt.Sprintf("@%d", m.contextCount))
		testContent := strings.Join(segments, sep) + sep + ctxSeg
		if lipgloss.Width(testContent) < m.width-8 {
			segments = append(segments, ctxSeg)
		}
	}

	// 6. Context percent (if fits)
	if m.contextPercent > 0 {
		pctSeg := mutedStyle.Render(fmt.Sprintf("%d%%", m.contextPercent))
		testContent := strings.Join(segments, sep) + sep + pctSeg
		if lipgloss.Width(testContent) < m.width-3 {
			segments = append(segments, pctSeg)
		}
	}

	// Join all segments
	content := strings.Join(segments, sep)
	contentWidth := lipgloss.Width(content)

	// Fill remaining space with diagonals
	leftDiags := diagStyle.Render(diag + diag + diag)
	rightPadding := m.width - contentWidth - 4 // 3 left diags + 1 space
	if rightPadding < 3 {
		rightPadding = 3
	}
	rightDiags := diagStyle.Render(strings.Repeat(diag, rightPadding))

	return leftDiags + " " + content + " " + rightDiags
}

// renderBrand renders "TERO" with optional sync status.
func (m *Model) renderBrand() string {
	colors := m.theme.Colors

	brand := styles.ApplyBoldForegroundGrad("TERO", colors.Brand.GradientStart, colors.Brand.GradientEnd)

	syncView := m.syncStatus.View()
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
