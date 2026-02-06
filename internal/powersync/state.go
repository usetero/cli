package powersync

import (
	"fmt"
)

// State represents the current syncer state.
// Use a type switch to handle each phase:
//
//	switch s := syncer.State().(type) {
//	case *powersync.Disconnected:
//	case *powersync.Connecting:
//	case *powersync.Syncing:
//	case *powersync.Ready:
//	case *powersync.Reconnecting:
//	case *powersync.Error:
//	}
type State interface {
	state() // marker method
}

// Disconnected means the syncer is not running.
type Disconnected struct{}

func (*Disconnected) state() {}

func NewDisconnected() *Disconnected {
	return &Disconnected{}
}

// Connecting means the syncer is establishing its first connection.
type Connecting struct{}

func (*Connecting) state() {}

func NewConnecting() *Connecting {
	return &Connecting{}
}

// Syncing means the syncer is actively downloading data.
type Syncing struct {
	Progress *Progress // nil until progress is known
}

func (*Syncing) state() {}

func NewSyncing() *Syncing {
	return &Syncing{}
}

func (s *Syncing) WithProgress(downloaded, total int) *Syncing {
	return &Syncing{
		Progress: &Progress{Downloaded: downloaded, Total: total},
	}
}

// Ready means the initial sync is complete and data is fresh.
type Ready struct{}

func (*Ready) state() {}

func NewReady() *Ready {
	return &Ready{}
}

// Reconnecting means the syncer lost its connection and is retrying.
// Degraded is true after repeated consecutive failures.
type Reconnecting struct {
	Degraded bool
}

func (*Reconnecting) state() {}

func NewReconnecting(degraded bool) *Reconnecting {
	return &Reconnecting{Degraded: degraded}
}

// Error means a fatal error occurred and syncing stopped.
type Error struct {
	Err error
}

func (*Error) state() {}

func NewError(err error) *Error {
	return &Error{Err: err}
}

// Progress represents download progress.
type Progress struct {
	Downloaded int
	Total      int
}

// String returns a human-readable progress string like "50/100".
func (p *Progress) String() string {
	return fmt.Sprintf("%d/%d", p.Downloaded, p.Total)
}
