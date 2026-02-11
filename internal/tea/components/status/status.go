// Package status renders colored status badges for Tero entities.
//
// Each entity type has its own renderer with the correct status-to-color
// mapping. All renderers produce a colored dot (●) followed by an optional
// label.
package status

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// badge renders ● LABEL or just ● depending on showLabel.
func badge(c color.Color, bg color.Color, label string, showLabel bool) string {
	style := lipgloss.NewStyle().Foreground(c).Background(bg)
	if showLabel {
		return style.Render("● " + label)
	}
	return style.Render("●")
}

// --- Service statuses: DISABLED > INACTIVE > BROKEN > STALE > DISCOVERING > ANALYZING > READY ---

// Service renders a colored status badge for a service status.
func Service(theme styles.Theme, s domain.ServiceLogStatus, showLabel bool) string {
	return badge(serviceColor(theme, s), theme.Bg, s.String(), showLabel)
}

// ServiceDot renders just the colored dot for a service status.
func ServiceDot(theme styles.Theme, s domain.ServiceLogStatus) string {
	return badge(serviceColor(theme, s), theme.Bg, "", false)
}

func serviceColor(theme styles.Theme, s domain.ServiceLogStatus) color.Color {
	switch s {
	case domain.ServiceLogStatusBroken, domain.ServiceLogStatusDisabled, domain.ServiceLogStatusInactive:
		return theme.Error
	case domain.ServiceLogStatusStale, domain.ServiceLogStatusDiscovering, domain.ServiceLogStatusAnalyzing:
		return theme.Warning
	case domain.ServiceLogStatusReady:
		return theme.Success
	default:
		return theme.TextMuted
	}
}

// --- Log event statuses: BROKEN > RESOLVED > CLEAN > PENDING > ANALYZING > DISCOVERING ---

// LogEvent renders a colored status badge for a log event status.
func LogEvent(theme styles.Theme, s domain.LogEventStatus, showLabel bool) string {
	return badge(logEventColor(theme, s), theme.Bg, s.String(), showLabel)
}

// LogEventDot renders just the colored dot for a log event status.
func LogEventDot(theme styles.Theme, s domain.LogEventStatus) string {
	return badge(logEventColor(theme, s), theme.Bg, "", false)
}

func logEventColor(theme styles.Theme, s domain.LogEventStatus) color.Color {
	switch s {
	case domain.LogEventStatusBroken, domain.LogEventStatusQuarantined:
		return theme.Error
	case domain.LogEventStatusResolved, domain.LogEventStatusClean:
		return theme.Success
	case domain.LogEventStatusPending, domain.LogEventStatusAnalyzing, domain.LogEventStatusDiscovering:
		return theme.Warning
	default:
		return theme.TextMuted
	}
}

// --- Waste badge: ● N% ---

// Waste renders a colored dot with a waste percentage.
// Returns empty string if pct is 0.
func Waste(theme styles.Theme, pct int) string {
	if pct <= 0 {
		return ""
	}
	dot := lipgloss.NewStyle().Foreground(theme.Warning).Background(theme.Bg).Render("●")
	text := lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Bg).Render(fmt.Sprintf(" %d%% waste", pct))
	return dot + text
}

// WasteShort renders a colored dot with just the percentage (no "waste" label).
// Useful for table cells.
func WasteShort(theme styles.Theme, pct int) string {
	if pct <= 0 {
		return ""
	}
	dot := lipgloss.NewStyle().Foreground(theme.Warning).Background(theme.Bg).Render("●")
	text := lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Bg).Render(fmt.Sprintf(" %d%%", pct))
	return dot + text
}

// --- Risk levels: HIGH > MEDIUM > LOW ---

// Risk renders a colored status badge for a risk level.
func Risk(theme styles.Theme, r domain.RiskLevel, showLabel bool) string {
	return badge(riskColor(theme, r), theme.Bg, r.String(), showLabel)
}

func riskColor(theme styles.Theme, r domain.RiskLevel) color.Color {
	switch r {
	case domain.RiskLevelHigh:
		return theme.Error
	case domain.RiskLevelMedium:
		return theme.Warning
	case domain.RiskLevelLow:
		return theme.Success
	default:
		return theme.TextMuted
	}
}

// --- Policy statuses: PENDING > APPROVED > DISMISSED ---

// Policy renders a colored status badge for a policy status.
func Policy(theme styles.Theme, s domain.PolicyLogStatus, showLabel bool) string {
	return badge(policyColor(theme, s), theme.Bg, s.String(), showLabel)
}

// PolicyDot renders just the colored dot for a policy status.
func PolicyDot(theme styles.Theme, s domain.PolicyLogStatus) string {
	return badge(policyColor(theme, s), theme.Bg, "", false)
}

func policyColor(theme styles.Theme, s domain.PolicyLogStatus) color.Color {
	switch s {
	case domain.PolicyLogStatusApproved:
		return theme.Success
	case domain.PolicyLogStatusPending:
		return theme.Warning
	case domain.PolicyLogStatusDismissed:
		return theme.TextMuted
	default:
		return theme.TextMuted
	}
}
