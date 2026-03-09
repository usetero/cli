package statusbar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/chrome"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

const productionEnv = "prd"

// StatusReader exposes runtime session status to the status bar.
type StatusReader interface {
	Status() sessionruntime.Status
}

// Model renders the global status bar header.
type Model struct {
	session StatusReader
	env     string
	theme   theme.Theme
	width   int
}

// New constructs a status bar model.
func New(session StatusReader, env string, appTheme theme.Theme) *Model {
	if session == nil {
		panic("status bar session runtime is required")
	}
	return &Model{
		session: session,
		env:     strings.ToLower(strings.TrimSpace(env)),
		theme:   appTheme,
	}
}

// SetWidth sets the available width for status bar rendering.
func (m *Model) SetWidth(width int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the full status bar line.
func (m *Model) View() string {
	status := m.session.Status()
	candidates := m.candidates(status)

	if m.width <= 0 {
		return candidates[0]
	}
	for _, candidate := range candidates {
		if lipgloss.Width(candidate) <= m.width {
			return candidate
		}
	}

	syncCompact := presentSync(status, true)
	if hasContext(status) {
		contextLabel := truncateLabel(presentContext(status, true), m.width-lipgloss.Width(syncCompact.icon)-1)
		return presentSyncContextLabel(m.theme, syncCompact, contextLabel)
	}
	return syncCompact.renderDot(m.theme)
}

func (m *Model) candidates(status sessionruntime.Status) []string {
	if hasContext(status) {
		if m.width > 48 || m.width <= 0 {
			return []string{
				m.renderCandidate(status, true, false, true, false),
				m.renderCandidate(status, false, false, false, false),
				m.renderCandidate(status, false, false, true, false),
				m.renderCandidate(status, false, true, false, false),
				m.renderCandidate(status, false, true, true, false),
			}
		}
		return []string{
			m.renderCandidate(status, false, false, false, false),
			m.renderCandidate(status, false, true, false, false),
			m.renderCandidate(status, true, false, true, false),
			m.renderCandidate(status, false, false, true, false),
			m.renderCandidate(status, false, true, true, false),
		}
	}

	return []string{
		m.renderCandidate(status, true, false, true, false),
		m.renderCandidate(status, false, false, true, false),
		m.renderCandidate(status, false, false, false, false),
		m.renderCandidate(status, false, true, false, false),
		m.renderCandidate(status, false, true, true, false),
		m.renderCandidate(status, false, false, true, true),
		m.renderCandidate(status, false, false, false, true),
		m.renderCandidate(status, false, true, false, true),
	}
}

func (m *Model) renderCandidate(status sessionruntime.Status, includeEnv bool, compactContext bool, includeHint bool, textualSync bool) string {
	left := m.renderLeft(status, includeEnv, compactContext, textualSync)
	right := ""
	if includeHint {
		right = m.renderDrawerHint()
	}
	return m.composeLine(left, right)
}

func (m *Model) renderLeft(status sessionruntime.Status, includeEnv bool, compactContext bool, textualSync bool) string {
	segments := []string{m.renderBrand()}
	statusSegment := m.renderStatusContext(status, compactContext, textualSync)
	if statusSegment != "" {
		segments = append(segments, statusSegment)
	}
	envSegment := ""
	if !hasContext(status) {
		envSegment = m.renderEnv(includeEnv)
	}
	if envSegment != "" {
		segments = append(segments, envSegment)
	}
	return strings.Join(segments, " ")
}

func (m *Model) renderStatusContext(status sessionruntime.Status, compactContext bool, textualSync bool) string {
	sync := presentSync(status, compactContext)
	if textualSync {
		return sync.render(m.theme)
	}

	return presentSyncContext(m.theme, status, compactContext)
}

func (m *Model) composeLine(left string, right string) string {
	leftDiags := chrome.RenderSlashMotif(m.theme, 2)
	rightDiags := chrome.RenderSlashMotif(m.theme, 2)

	if m.width <= 0 {
		if right == "" {
			return lipgloss.JoinHorizontal(lipgloss.Left, leftDiags, " ", left)
		}
		return lipgloss.JoinHorizontal(lipgloss.Left, leftDiags, " ", left, " ", right, " ", rightDiags)
	}

	leftWidth := lipgloss.Width(left)
	leftDiagsWidth := lipgloss.Width(leftDiags)
	if right == "" {
		motifWidth := m.width - leftDiagsWidth - leftWidth - 2
		if motifWidth <= 0 {
			return lipgloss.JoinHorizontal(lipgloss.Left, leftDiags, " ", left)
		}
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftDiags,
			" ",
			left,
			" ",
			chrome.RenderSlashMotif(m.theme, motifWidth),
		)
	}

	rightWidth := lipgloss.Width(right)
	rightDiagsWidth := lipgloss.Width(rightDiags)
	motifWidth := m.width - leftDiagsWidth - leftWidth - rightWidth - rightDiagsWidth - 4
	if motifWidth <= 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, leftDiags, " ", left, " ", right, " ", rightDiags)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftDiags,
		" ",
		left,
		" ",
		chrome.RenderSlashMotif(m.theme, motifWidth),
		" ",
		right,
		" ",
		rightDiags,
	)
}

func (m *Model) renderDrawerHint() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.theme.Shell.HeaderLead.Render("ctrl+d"),
		m.theme.Shell.HeaderRule.Render(" open"),
	)
}

func truncateLabel(label string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(label)
	if len(runes) <= width {
		return label
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}
