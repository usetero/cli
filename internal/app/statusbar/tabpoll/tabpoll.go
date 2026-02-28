package tabpoll

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/sqlite"
)

// PollMsg triggers a tab data refresh tick.
type PollMsg struct{}

// DataMsg carries typed async fetch results back to a tab model.
type DataMsg[T any] struct {
	Data T
	Err  error
}

// Tick schedules the next poll tick.
func Tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return PollMsg{}
	})
}

// Fetch executes a typed data fetch with a timeout and returns DataMsg[T].
func Fetch[T any](timeout time.Duration, fetch func(ctx context.Context) (T, error)) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := sqlite.WithTimeout(context.Background(), timeout)
		defer cancel()
		data, err := fetch(ctx)
		return DataMsg[T]{Data: data, Err: err}
	}
}

// ApplyIfChanged applies an update only when the state key changes and clamps cursor.
func ApplyIfChanged(lastState *string, nextState string, cursor *int, length int, apply func()) bool {
	if *lastState == nextState {
		return false
	}
	apply()
	*lastState = nextState
	if cursor != nil {
		*cursor = listdetail.ClampCursor(*cursor, length)
	}
	return true
}
