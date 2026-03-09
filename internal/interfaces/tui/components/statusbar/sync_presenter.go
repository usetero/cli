package statusbar

import (
	"fmt"

	"charm.land/lipgloss/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type syncTone int

const (
	syncToneMuted syncTone = iota
	syncToneWarning
	syncToneSuccess
	syncToneError
)

type syncPresentation struct {
	icon  string
	label string
	tone  syncTone
}

func presentSync(status sessionruntime.Status, compact bool) syncPresentation {
	if !status.Running {
		if compact {
			return syncPresentation{icon: "●", label: "off", tone: syncToneError}
		}
		return syncPresentation{icon: "●", label: "offline", tone: syncToneError}
	}

	switch state := status.Sync.(type) {
	case *pssyncer.Ready:
		return syncPresentation{icon: "●", label: "ready", tone: syncToneSuccess}
	case *pssyncer.Error:
		if compact {
			return syncPresentation{icon: "●", label: "err", tone: syncToneError}
		}
		return syncPresentation{icon: "●", label: "error", tone: syncToneError}
	case *pssyncer.Connecting:
		if compact {
			return syncPresentation{icon: "●", label: "conn", tone: syncToneWarning}
		}
		return syncPresentation{icon: "●", label: "connecting", tone: syncToneWarning}
	case *pssyncer.Reconnecting:
		if compact {
			return syncPresentation{icon: "●", label: "reconn", tone: syncToneWarning}
		}
		return syncPresentation{icon: "●", label: "reconnecting", tone: syncToneWarning}
	case *pssyncer.Syncing:
		if state.Progress != nil && state.Progress.Total > 0 {
			return syncPresentation{
				icon:  "●",
				label: fmt.Sprintf("sync %d/%d", state.Progress.Downloaded, state.Progress.Total),
				tone:  syncToneWarning,
			}
		}
		if compact {
			return syncPresentation{icon: "●", label: "sync", tone: syncToneWarning}
		}
		return syncPresentation{icon: "●", label: "syncing", tone: syncToneWarning}
	case *pssyncer.Disconnected:
		if compact {
			return syncPresentation{icon: "●", label: "disc", tone: syncToneError}
		}
		return syncPresentation{icon: "●", label: "disconnected", tone: syncToneError}
	default:
		if compact {
			return syncPresentation{icon: "●", label: "unk", tone: syncToneError}
		}
		return syncPresentation{icon: "●", label: "unknown", tone: syncToneError}
	}
}

func (s syncPresentation) render(t theme.Theme) string {
	return s.renderWithLabel(t, s.label)
}

func (s syncPresentation) renderWithLabel(t theme.Theme, label string) string {
	value := s.icon
	if label != "" {
		value += " " + label
	}
	return syncToneStyle(t, s.tone).Render(value)
}

func (s syncPresentation) renderDot(t theme.Theme) string {
	return syncToneStyle(t, s.tone).Render(s.icon)
}

func syncToneStyle(t theme.Theme, tone syncTone) lipgloss.Style {
	switch tone {
	case syncToneSuccess:
		return t.Text.Success
	case syncToneWarning:
		return t.Text.Warning
	case syncToneError:
		return t.Text.Error
	default:
		return t.Text.Muted
	}
}
