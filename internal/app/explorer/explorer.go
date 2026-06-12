// Package explorer renders a minimal, read-only view of the account's active
// issues. It is the default interactive surface now that chat is gone.
package explorer

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

const fetchTimeout = 5 * time.Second

var (
	keyUp      = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑↓", "navigate"))
	keyDown    = key.NewBinding(key.WithKeys("down", "j"))
	keyRefresh = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh"))
)

// Model is the issue explorer.
type Model struct {
	services graphql.ServiceSet
	scope    log.Scope
	theme    styles.Theme

	width, height    int
	originX, originY int

	issues  []domain.Issue
	summary domain.IssueSummary
	cursor  int
	loading bool
	err     error
}

type issuesLoadedMsg struct {
	issues  []domain.Issue
	summary domain.IssueSummary
	err     error
}

// New creates a new explorer model.
func New(services graphql.ServiceSet, theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		services: services,
		theme:    theme,
		scope:    scope.Child("explorer"),
		loading:  true,
	}
}

// Init fetches the initial issue list.
func (m *Model) Init() tea.Cmd {
	return m.fetch()
}

// Refresh re-fetches the issue list.
func (m *Model) Refresh() tea.Cmd {
	m.loading = true
	return m.fetch()
}

func (m *Model) fetch() tea.Cmd {
	services := m.services
	scope := m.scope
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		issues, err := services.Issues.List(ctx)
		if err != nil {
			scope.Error("list issues", "err", err)
			return issuesLoadedMsg{err: err}
		}
		summary, err := services.Issues.GetSummary(ctx)
		if err != nil {
			scope.Error("issue summary", "err", err)
			return issuesLoadedMsg{issues: issues, err: err}
		}
		return issuesLoadedMsg{issues: issues, summary: summary}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case issuesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.issues = msg.issues
		m.summary = msg.summary
		if m.cursor >= len(m.issues) {
			m.cursor = max(0, len(m.issues)-1)
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keyDown):
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		case key.Matches(msg, keyRefresh):
			return m.Refresh()
		}
	}
	return nil
}

// SetSize updates the dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetOrigin records the page origin (kept for layout parity with other pages).
func (m *Model) SetOrigin(x, y int) {
	m.originX = x
	m.originY = y
}

// ShortHelp returns the key bindings shown in the keybar.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keyUp, keyRefresh}
}

// View renders the explorer.
func (m *Model) View() string {
	colors := m.theme
	title := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg).Bold(true)
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	var b strings.Builder
	b.WriteString(title.Render("Active issues"))
	b.WriteString("  ")
	b.WriteString(muted.Render(m.headline()))
	b.WriteString("\n\n")

	switch {
	case m.loading:
		b.WriteString(muted.Render("Loading issues…"))
		return b.String()
	case m.err != nil:
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("Failed to load issues: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("Press r to retry."))
		return b.String()
	case len(m.issues) == 0:
		b.WriteString(muted.Render("No active issues. 🎉"))
		return b.String()
	}

	for i, issue := range m.issues {
		b.WriteString(m.renderRow(i, issue))
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) headline() string {
	if m.loading || m.err != nil {
		return ""
	}
	return fmt.Sprintf("%d open · %d high · %d medium · %d low",
		m.summary.Open, m.summary.HighCount, m.summary.MediumCount, m.summary.LowCount)
}

func (m *Model) renderRow(index int, issue domain.Issue) string {
	colors := m.theme
	cursor := "  "
	nameStyle := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	if index == m.cursor {
		cursor = lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg).Render("▶ ")
		nameStyle = nameStyle.Foreground(colors.Accent)
	}

	prio := lipgloss.NewStyle().Background(colors.Bg).Foreground(priorityColor(colors, issue.Priority)).Render(pad(string(issue.Priority), 6))
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	meta := issue.DisplayID
	if issue.ServiceName != "" {
		meta += " · " + issue.ServiceName
	}

	line := cursor + prio + "  " + nameStyle.Render(truncate(issue.Title, m.titleWidth())) + "  " + muted.Render(meta)
	return line
}

func (m *Model) titleWidth() int {
	w := m.width - 30
	if w < 20 {
		w = 20
	}
	return w
}

func priorityColor(theme styles.Theme, p domain.IssuePriority) color.Color {
	switch p {
	case domain.IssuePriorityHigh:
		return theme.Error
	case domain.IssuePriorityMedium:
		return theme.Warning
	default:
		return theme.TextMuted
	}
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
	if n <= 1 || len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
