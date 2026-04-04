package commandbar

// Mode is the active interaction state for the global command bar.
type Mode int

const (
	ModeHidden Mode = iota
	ModeError
	ModeAction
	ModeInput
	ModeSelect
)

// SurfaceState is the shell-owned presentation state for the command bar.
type SurfaceState int

const (
	SurfaceHidden SurfaceState = iota
	SurfaceActive
	SurfaceBusy
	SurfaceError
)

func (m *Model) surfaceState() SurfaceState {
	if m.err != nil {
		return SurfaceError
	}
	if m.busy != nil {
		return SurfaceBusy
	}
	switch m.mode {
	case ModeAction, ModeInput, ModeSelect:
		return SurfaceActive
	default:
		return SurfaceHidden
	}
}
