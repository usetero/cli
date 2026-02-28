package viewkit

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// RenderPolicyEmptyState renders the common empty/loading/disabled/healthy
// states for policy-driven status tabs (waste/quality/compliance).
func RenderPolicyEmptyState(
	theme styles.Theme,
	dbReady bool,
	summary domain.AccountSummary,
	disabledHint string,
	healthyText string,
) string {
	muted := lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.Bg)
	if !dbReady {
		return muted.Render("Waiting for sync to start...")
	}
	if summary.ActiveServices == 0 && summary.ServiceCount > 0 {
		return muted.Render(fmt.Sprintf(
			"%d services discovered, all disabled.\n%s",
			summary.ServiceCount,
			disabledHint,
		))
	}
	if summary.ActiveServices == 0 {
		return muted.Render("No services discovered yet.")
	}
	dot := lipgloss.NewStyle().Foreground(theme.Success).Background(theme.Bg).Render("●")
	return dot + " " + muted.Render(healthyText)
}

// ComposeSummaryTableView joins a headline + table + optional description
// in the standard expanded-tab layout.
func ComposeSummaryTableView(theme styles.Theme, headline, tableView, description string) string {
	var lines []string
	lines = append(lines, headline)
	lines = append(lines, "")

	if tableView != "" {
		lines = append(lines, tableView)
	}

	if description != "" {
		lines = append(lines, "")
		muted := lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.Bg)
		lines = append(lines, muted.Render(description))
	}

	return strings.Join(lines, "\n")
}
