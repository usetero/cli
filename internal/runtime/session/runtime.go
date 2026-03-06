package session

import "github.com/usetero/cli/internal/infrastructure/logging"

// Runtime is the account session runtime lifecycle type.
type Runtime = Service

// NewRuntime constructs the account session runtime.
func NewRuntime(storage Storage, newSyncer syncerFactory, newUploader uploaderFactory, log logging.Scope) *Runtime {
	return NewService(storage, newSyncer, newUploader, log)
}
