package account

import (
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Status is the UI-facing snapshot of current runtime lifecycle + sync state.
type Status struct {
	Running                 bool
	Scope                   Scope
	Ready                   bool
	SessionReady            bool
	HasCompletedInitialSync bool
	Sync                    pssyncer.State
	DBPath                  sqlite.DatabasePath
}

func (r *Runtime) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()

	status := Status{
		Sync: &pssyncer.Disconnected{},
	}
	if r.closed || r.db == nil || r.syncer == nil {
		return status
	}

	status.Running = true
	status.Scope = r.scope
	status.HasCompletedInitialSync = r.hasCompletedInitialSync
	status.SessionReady = r.syncer.IsReady()
	status.Ready = status.SessionReady
	status.Sync = r.syncer.State()
	status.DBPath = r.dbPath
	if status.Sync == nil {
		status.Sync = &pssyncer.Disconnected{}
	}
	return status
}
