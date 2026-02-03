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
//	case *powersync.Error:
//	}
type State interface {
	state() // marker method
}

// Disconnected means the syncer is not running.
type Disconnected struct{}

func (*Disconnected) state() {}

// NewDisconnected creates a disconnected state.
func NewDisconnected() *Disconnected {
	return &Disconnected{}
}

// Connecting means the syncer is establishing a connection.
type Connecting struct {
	Message string
}

func (*Connecting) state() {}

// NewConnecting creates a connecting state.
func NewConnecting(message string) *Connecting {
	return &Connecting{Message: message}
}

// Syncing means the syncer is actively downloading data.
type Syncing struct {
	Message  string    // Always set: "Syncing your data..." or with progress
	Warning  string    // Transient issue, empty if none
	Progress *Progress // Download progress, nil if not yet known
}

func (*Syncing) state() {}

// NewSyncing creates a syncing state with the given message.
func NewSyncing(message string) *Syncing {
	return &Syncing{Message: message}
}

// WithProgress returns a copy with progress set.
func (s *Syncing) WithProgress(downloaded, total int) *Syncing {
	return &Syncing{
		Message:  s.Message,
		Warning:  s.Warning,
		Progress: &Progress{Downloaded: downloaded, Total: total},
	}
}

// UpdateProgress returns a copy with updated progress and message, preserving warning.
func (s *Syncing) UpdateProgress(downloaded, total int) *Syncing {
	return &Syncing{
		Message:  fmt.Sprintf("Syncing your data... (%d/%d)", downloaded, total),
		Warning:  s.Warning,
		Progress: &Progress{Downloaded: downloaded, Total: total},
	}
}

// WithWarning returns a copy with warning set.
func (s *Syncing) WithWarning(warning string) *Syncing {
	return &Syncing{
		Message:  s.Message,
		Warning:  warning,
		Progress: s.Progress,
	}
}

// Ready means the initial sync is complete.
type Ready struct{}

func (*Ready) state() {}

// NewReady creates a ready state.
func NewReady() *Ready {
	return &Ready{}
}

// Error means a fatal error occurred and syncing stopped.
type Error struct {
	Err error
}

func (*Error) state() {}

// NewError creates an error state.
func NewError(err error) *Error {
	return &Error{Err: err}
}

// Progress represents download progress.
type Progress struct {
	Downloaded int
	Total      int
}
