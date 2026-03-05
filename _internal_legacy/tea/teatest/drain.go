package teatest

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// DrainCmds executes a tea.Cmd and feeds resulting messages back through
// the update function. Handles tea.BatchMsg by expanding and executing each
// sub-command.
//
// Each cmd is executed with a timeout. Cmds that block (e.g. channel reads
// waiting on a stream) are skipped. This makes DrainCmds safe to use with
// models that have in-flight async work.
//
// Stops after maxSteps to prevent infinite loops in tests.
func DrainCmds(update func(tea.Msg) tea.Cmd, cmd tea.Cmd, maxSteps int) {
	if cmd == nil {
		return
	}
	queue := []tea.Cmd{cmd}

	for step := 0; step < maxSteps && len(queue) > 0; step++ {
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}

		msg, ok := execCmd(c, 200*time.Millisecond)
		if !ok || msg == nil {
			continue
		}

		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, batch...)
			continue
		}

		if next := update(msg); next != nil {
			queue = append(queue, next)
		}
	}
}

// execCmd runs a tea.Cmd with a timeout. Returns the message and true if the
// cmd completed, or nil and false if it blocked.
func execCmd(cmd tea.Cmd, timeout time.Duration) (tea.Msg, bool) {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()

	select {
	case msg := <-ch:
		return msg, true
	case <-time.After(timeout):
		return nil, false
	}
}
