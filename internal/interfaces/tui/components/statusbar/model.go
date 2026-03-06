package statusbar

import (
	"strings"

	"charm.land/lipgloss/v2"
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
	syncFull := presentSync(status, false)
	syncCompact := presentSync(status, true)

	candidates := []string{
		m.renderLine(true, syncFull, m.separator()),
		m.renderLine(false, syncFull, m.separator()),
		m.renderLine(false, syncCompact, m.separator()),
		m.renderLine(false, syncCompact, " "),
		syncCompact.render(m.theme),
	}

	if m.width <= 0 {
		return candidates[0]
	}
	for _, candidate := range candidates {
		if lipgloss.Width(candidate) <= m.width {
			return candidate
		}
	}

	truncated := truncateLabel(syncCompact.label, m.width)
	return syncCompact.renderWithLabel(
		m.theme,
		truncated,
	)
}

func (m *Model) separator() string {
	return m.theme.Shell.HeaderLead.Render("  ╱  ")
}

func (m *Model) renderLine(includeEnv bool, sync syncPresentation, separator string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.renderBrand(includeEnv),
		separator,
		sync.render(m.theme),
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
