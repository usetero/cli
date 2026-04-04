package understanding

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/catalog"
)

// SQLiteReader is the SQLite-backed Understanding read model implementation.
//
// Query wiring will land here once the initial SQL shape is defined.
type SQLiteReader struct{}

func NewSQLiteReader() *SQLiteReader {
	return &SQLiteReader{}
}

func (r *SQLiteReader) Snapshot(ctx context.Context, req SnapshotRequest) (Snapshot, error) {
	return Snapshot{}, fmt.Errorf("understanding snapshot query not implemented")
}

func (r *SQLiteReader) EventDetail(ctx context.Context, id catalog.LogEventID) (EventDetail, error) {
	return EventDetail{}, fmt.Errorf("understanding event detail query not implemented")
}
