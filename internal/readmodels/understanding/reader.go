package understanding

import (
	"context"

	"github.com/usetero/cli/internal/domains/catalog"
)

// Reader exposes the Understanding read model.
type Reader interface {
	Snapshot(ctx context.Context, req SnapshotRequest) (Snapshot, error)
	EventDetail(ctx context.Context, id catalog.LogEventID) (EventDetail, error)
}
