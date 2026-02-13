// Package confirm provides the confirmation dialog step for the policy approval wizard.
package confirm

import (
	"fmt"
	"math"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/policyapproval/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model handles the confirmation dialog — the third step of the approval wizard.
type Model struct {
	theme           styles.Theme
	scope           log.Scope
	category        string
	policies        []domain.PolicyDetail
	selectedApprove bool // which button is focused (true = Approve, false = Cancel)
	width           int
	height          int
}

// New creates a new confirm step.
func New(theme styles.Theme, category string, policies []domain.PolicyDetail, scope log.Scope) *Model {
	return &Model{
		theme:           theme,
		scope:           scope.Child("confirm"),
		category:        category,
		policies:        policies,
		selectedApprove: true, // default to Approve
	}
}

// Init returns nil — no async work needed.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("confirm step started", "category", m.category, "count", len(m.policies))
	return nil
}

// Update handles key presses.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, keyToggle):
			m.selectedApprove = !m.selectedApprove
		case key.Matches(msg, keyConfirm):
			if m.selectedApprove {
				m.scope.Info("approval confirmed")
				return func() tea.Msg { return msgs.Confirmed{} }
			}
			m.scope.Info("approval cancelled from confirm dialog")
			return func() tea.Msg { return msgs.BackToSummary{} }
		case key.Matches(msg, keyCancel):
			m.scope.Info("confirm cancelled via esc")
			return func() tea.Msg { return msgs.BackToSummary{} }
		}
	}
	return nil
}

// View renders the confirmation dialog.
func (m *Model) View() string {
	colors := m.theme
	count := len(m.policies)
	savings := m.estimatedSavings()

	// Title
	title := lipgloss.NewStyle().
		Foreground(colors.Text).
		Bold(true).
		Render(fmt.Sprintf("Approve %d policies in %s?", count, m.category))

	// Description
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted)
	var details []string
	details = append(details, muted.Render("This will create Datadog exclusion filters for:"))
	details = append(details, muted.Render(fmt.Sprintf("  • %d log events", count)))
	if savings > 0 {
		details = append(details, muted.Render(fmt.Sprintf("  • Estimated savings: ~%s/yr", formatCost(savings))))
	}
	description := strings.Join(details, "\n")

	// Buttons
	cancelBtn := m.renderButton("Cancel", !m.selectedApprove)
	approveBtn := m.renderButton("Approve", m.selectedApprove)
	buttons := cancelBtn + "    " + approveBtn

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", description, "", buttons)

	return lipgloss.NewStyle().
		Foreground(colors.Text).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Accent).
		Padding(1, 3).
		Render(content)
}

// renderButton renders a single button with selected/unselected styling.
func (m *Model) renderButton(label string, selected bool) string {
	colors := m.theme

	if selected {
		return lipgloss.NewStyle().
			Background(colors.Accent).
			Foreground(colors.Bg).
			Padding(0, 3).
			Bold(true).
			Render(label)
	}

	return lipgloss.NewStyle().
		Foreground(colors.TextMuted).
		Padding(0, 3).
		Render(label)
}

// estimatedSavings calculates total yearly savings for selected policies.
func (m *Model) estimatedSavings() float64 {
	var total float64
	for _, p := range m.policies {
		total += p.EstimatedCostPerHour
	}
	return total * 8760 // hours per year
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keyToggle, keyConfirm, keyCancel}
}

// Key bindings.
var (
	keyToggle  = key.NewBinding(key.WithKeys("tab", "left", "right"), key.WithHelp("tab", "switch"))
	keyConfirm = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))
	keyCancel  = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

// formatCost formats a dollar amount: $142, $9.4k, $1.2M.
func formatCost(dollars float64) string {
	abs := math.Abs(dollars)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("$%.1fM", dollars/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("$%.1fk", dollars/1_000)
	default:
		return fmt.Sprintf("$%.0f", dollars)
	}
}
