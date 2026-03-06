package uploader

import (
	"errors"
	"fmt"

	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

// ErrAlreadyRun reports that an uploader instance was already started once.
// Uploader instances are one-shot and should be recreated for a new run.
var ErrAlreadyRun = errors.New("uploader run already started")

// UnknownMutationHandlerError reports that a queued mutation table has no
// configured upload handler, so the uploader cannot safely advance the queue.
type UnknownMutationHandlerError struct {
	Table psdb.TableName
}

func (e UnknownMutationHandlerError) Error() string {
	return fmt.Sprintf("no upload handler configured for table %q", e.Table)
}
