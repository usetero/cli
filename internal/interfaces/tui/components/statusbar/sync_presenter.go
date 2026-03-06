package statusbar

import (
	"fmt"

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
	label string
	tone  syncTone
}

func presentSync(status sessionruntime.Status, compact bool) syncPresentation {
	if !status.Running {
		if compact {
			return syncPresentation{label: "○ off", tone: syncToneMuted}
		}
		return syncPresentation{label: "○ offline", tone: syncToneMuted}
	}

	switch state := status.Sync.(type) {
	case *pssyncer.Ready:
		return syncPresentation{label: "● ready", tone: syncToneSuccess}
	case *pssyncer.Error:
		if compact {
			return syncPresentation{label: "● err", tone: syncToneError}
		}
		return syncPresentation{label: "● error", tone: syncToneError}
	case *pssyncer.Connecting:
		if compact {
			return syncPresentation{label: "● conn", tone: syncToneWarning}
		}
		return syncPresentation{label: "● connecting", tone: syncToneWarning}
	case *pssyncer.Reconnecting:
		if compact {
			return syncPresentation{label: "● reconn", tone: syncToneWarning}
		}
		return syncPresentation{label: "● reconnecting", tone: syncToneWarning}
	case *pssyncer.Syncing:
		if state.Progress != nil && state.Progress.Total > 0 {
			return syncPresentation{
				label: fmt.Sprintf("● sync %d/%d", state.Progress.Downloaded, state.Progress.Total),
				tone:  syncToneWarning,
			}
		}
		if compact {
			return syncPresentation{label: "● sync", tone: syncToneWarning}
		}
		return syncPresentation{label: "● syncing", tone: syncToneWarning}
	case *pssyncer.Disconnected:
		if compact {
			return syncPresentation{label: "○ disc", tone: syncToneMuted}
		}
		return syncPresentation{label: "○ disconnected", tone: syncToneMuted}
	default:
		if compact {
			return syncPresentation{label: "○ unk", tone: syncToneMuted}
		}
		return syncPresentation{label: "○ unknown", tone: syncToneMuted}
	}
}

func (s syncPresentation) render(t theme.Theme) string {
	return s.renderWithLabel(t, s.label)
}

func (s syncPresentation) renderWithLabel(t theme.Theme, label string) string {
	switch s.tone {
	case syncToneSuccess:
		return t.Text.Success.Render(label)
	case syncToneWarning:
		return t.Text.Warning.Render(label)
	case syncToneError:
		return t.Text.Error.Render(label)
	default:
		return t.Text.Muted.Render(label)
	}
}
