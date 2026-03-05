package syncer

import "fmt"

// State describes the current sync lifecycle phase.
type State interface {
	state()
}

// Disconnected means sync is not running.
type Disconnected struct{}

func (*Disconnected) state() {}

// Connecting means sync startup/control-plane initialization is in progress.
type Connecting struct{}

func (*Connecting) state() {}

// Syncing means stream is connected and downloading data.
type Syncing struct {
	Progress *Progress
}

func (*Syncing) state() {}

// Ready means initial sync has completed at least once.
type Ready struct{}

func (*Ready) state() {}

// Reconnecting means sync is retrying after an interruption.
type Reconnecting struct {
	Degraded bool
}

func (*Reconnecting) state() {}

// Error means sync stopped after a permanent failure.
type Error struct {
	Err error
}

func (*Error) state() {}

// Progress summarizes download progress.
type Progress struct {
	Downloaded int
	Total      int
}

func (p *Progress) String() string {
	if p == nil {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", p.Downloaded, p.Total)
}
