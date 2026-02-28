package statusbar

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
)

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

	// 4. Quality status (field-level improvements)
	qualityView := m.qualityStatus.CompactView()
	if qualityView != "" {
		segments = append(segments, qualityView)
	}

	// 5. Compliance status (leaking/at-risk counts across PII, Secrets, PHI, Payment Data)
	complianceView := m.complianceStatus.CompactView()
	if complianceView != "" {
		segments = append(segments, complianceView)
	}

	// Build right-aligned segment first so we know how much space is left.
	rightSeg := m.renderDrawerHint()
	rightWidth := lipgloss.Width(rightSeg)

	// Reserved: left diags (2) + spaces (3) + right diags (2) + right segment + min middle diags (3)
	reserved := 7 + rightWidth + 3
	if rightSeg == "" {
		reserved = 3 + 1 // left diags (2) + space + trailing min (1)
	}

	// Calculate what fits
	sepWidth := lipgloss.Width(sep)
	baseContent := strings.Join(segments, sep)
	baseWidth := lipgloss.Width(baseContent)

	// 6. Title (if fits)
	if m.title != "" {
		maxTitle := m.width - baseWidth - sepWidth - reserved
		titleSeg := m.renderTitle(maxTitle)
		if titleSeg != "" {
			segments = append(segments, titleSeg)
		}
	}

	// 7. Context window usage (only shown when high)
	if m.contextPercent >= 75 {
		pctSeg := m.renderContextPercent()
		testWidth := lipgloss.Width(strings.Join(segments, sep)) + sepWidth + lipgloss.Width(pctSeg)
		if testWidth < m.width-reserved {
			segments = append(segments, pctSeg)
		}
	}

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
	if tab := m.activeTabModel(); tab != nil {
		content = tab.ExpandedView(contentWidth, contentHeight)
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
	for i, tab := range m.tabs {
		label := tab.Label()
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
	if !m.drawerOpen && !m.anyTabHasData() {
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
