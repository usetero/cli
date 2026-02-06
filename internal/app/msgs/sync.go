package msgs

import "github.com/usetero/cli/internal/powersync"

// SyncStateChanged is emitted when sync state changes.
type SyncStateChanged struct {
	State powersync.State
}
