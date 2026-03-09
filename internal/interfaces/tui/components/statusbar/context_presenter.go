package statusbar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

func presentContext(status sessionruntime.Status, compact bool) string {
	org := displayValue(status.Scope.Organization.Name, string(status.Scope.Organization.ID))
	if org == "" {
		return ""
	}
	return org
}

func hasContext(status sessionruntime.Status) bool {
	return presentContext(status, false) != ""
}

func presentSyncContext(t theme.Theme, status sessionruntime.Status, compact bool) string {
	sync := presentSync(status, compact)
	contextLabel := presentContext(status, compact)
	return presentSyncContextLabel(t, sync, contextLabel)
}

func presentSyncContextLabel(t theme.Theme, sync syncPresentation, contextLabel string) string {
	if contextLabel == "" {
		return sync.renderDot(t)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		sync.renderDot(t),
		" ",
		t.Shell.HeaderLead.Render(contextLabel),
	)
}

func displayValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
