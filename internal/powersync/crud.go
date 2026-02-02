package powersync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
)

// CRUD operation types.
const (
	OpPut    = "PUT"    // Insert or replace
	OpPatch  = "PATCH"  // Update
	OpDelete = "DELETE" // Delete
)

// CrudEntry represents a single entry in the ps_crud upload queue.
type CrudEntry struct {
	// ID is the auto-incrementing client-side id.
	ID int64
	// TxID is the transaction id. All operations in the same transaction share this.
	TxID *int64
	// Op is the operation type: PUT, PATCH, or DELETE.
	Op string
	// Table is the table name.
	Table string
	// RowID is the ID of the affected row.
	RowID string
	// Data contains the row data (for PUT/PATCH operations).
	Data map[string]any
	// Old contains previous values (for tables with trackPreviousValues enabled).
	Old map[string]any
	// Metadata contains client-side metadata (if trackMetadata was enabled).
	Metadata *string
}

// crudRow is the raw row from ps_crud.
type crudRow struct {
	ID   int64  `json:"id"`
	TxID *int64 `json:"tx_id"`
	Data string `json:"data"`
}

// crudData is the JSON structure inside the data column.
type crudData struct {
	Op       string         `json:"op"`
	Type     string         `json:"type"`
	ID       string         `json:"id"`
	Data     map[string]any `json:"data,omitempty"`
	Old      map[string]any `json:"old,omitempty"`
	Metadata *string        `json:"metadata,omitempty"`
}

// CrudQueue provides access to the PowerSync CRUD upload queue.
type CrudQueue struct {
	db sqlite.Database
}

// NewCrudQueue creates a new CRUD queue accessor.
func NewCrudQueue(db sqlite.Database) *CrudQueue {
	return &CrudQueue{
		db: db,
	}
}

// HasPendingUploads returns true if there are entries waiting to be uploaded.
func (q *CrudQueue) HasPendingUploads(ctx context.Context) (bool, error) {
	var count int64
	err := q.db.DB().QueryRow(ctx, "SELECT COUNT(*) FROM ps_crud").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check pending uploads: %w", err)
	}
	return count > 0, nil
}

// GetNextEntry returns the next CRUD entry to process, or nil if the queue is empty.
func (q *CrudQueue) GetNextEntry(ctx context.Context) (*CrudEntry, error) {
	var row crudRow
	err := q.db.DB().QueryRow(ctx, "SELECT id, tx_id, data FROM ps_crud ORDER BY id LIMIT 1").Scan(
		&row.ID, &row.TxID, &row.Data,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get next crud entry: %w", err)
	}

	return q.parseEntry(row)
}

// GetNextTransaction returns all CRUD entries for the next transaction.
// If entries exist without a transaction ID, returns just the first entry.
func (q *CrudQueue) GetNextTransaction(ctx context.Context) ([]CrudEntry, error) {
	// First, get the minimum ID to find where to start
	first, err := q.GetNextEntry(ctx)
	if err != nil || first == nil {
		return nil, err
	}

	// If no transaction ID, return just this entry
	if first.TxID == nil {
		return []CrudEntry{*first}, nil
	}

	// Get all entries with the same transaction ID
	rows, err := q.db.DB().Query(ctx,
		"SELECT id, tx_id, data FROM ps_crud WHERE tx_id = ? ORDER BY id",
		*first.TxID,
	)
	if err != nil {
		return nil, fmt.Errorf("get transaction entries: %w", err)
	}
	defer rows.Close()

	var entries []CrudEntry
	for rows.Next() {
		var row crudRow
		if err := rows.Scan(&row.ID, &row.TxID, &row.Data); err != nil {
			return nil, fmt.Errorf("scan crud row: %w", err)
		}
		entry, err := q.parseEntry(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}

	return entries, rows.Err()
}

// GetAllEntries returns all pending CRUD entries in order.
func (q *CrudQueue) GetAllEntries(ctx context.Context) ([]*CrudEntry, error) {
	rows, err := q.db.DB().Query(ctx, "SELECT id, tx_id, data FROM ps_crud ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query crud entries: %w", err)
	}
	defer rows.Close()

	var entries []*CrudEntry
	for rows.Next() {
		var row crudRow
		if err := rows.Scan(&row.ID, &row.TxID, &row.Data); err != nil {
			return nil, fmt.Errorf("scan crud row: %w", err)
		}
		entry, err := q.parseEntry(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// ErrDatabaseCorrupt indicates the database is in an inconsistent state
// from a crash during migration or write. The database should be deleted
// and re-synced from the server.
var ErrDatabaseCorrupt = fmt.Errorf("database corrupt")

// CheckHealth verifies the database is in a consistent state.
// Returns ErrDatabaseCorrupt if corruption is detected from a crash.
func (q *CrudQueue) CheckHealth(ctx context.Context) error {
	// Check 1: ps_tx must have a row with id=1
	// This is required for CRUD operations. Missing if crash during migration.
	var txCount int
	err := q.db.DB().QueryRow(ctx, "SELECT COUNT(*) FROM ps_tx WHERE id = 1").Scan(&txCount)
	if err != nil {
		return fmt.Errorf("check ps_tx: %w", err)
	}
	if txCount == 0 {
		return fmt.Errorf("%w: ps_tx missing required row", ErrDatabaseCorrupt)
	}

	// Check 2: $local bucket should not be stuck
	// Stuck = target_op > last_op with empty ps_crud (crash during upload)
	hasPending, err := q.HasPendingUploads(ctx)
	if err != nil {
		return fmt.Errorf("check pending uploads: %w", err)
	}
	if !hasPending {
		var targetOp, lastOp int64
		err = q.db.DB().QueryRow(ctx,
			"SELECT target_op, last_op FROM ps_buckets WHERE name = '$local'",
		).Scan(&targetOp, &lastOp)
		if err != nil && err.Error() != "sql: no rows in result set" {
			return fmt.Errorf("check local bucket: %w", err)
		}
		if err == nil && targetOp > lastOp {
			return fmt.Errorf("%w: $local bucket stuck", ErrDatabaseCorrupt)
		}
	}

	return nil
}

// parseEntry converts a raw database row into a CrudEntry.
func (q *CrudQueue) parseEntry(row crudRow) (*CrudEntry, error) {
	var data crudData
	if err := json.Unmarshal([]byte(row.Data), &data); err != nil {
		return nil, fmt.Errorf("parse crud data: %w", err)
	}

	return &CrudEntry{
		ID:       row.ID,
		TxID:     row.TxID,
		Op:       data.Op,
		Table:    data.Type,
		RowID:    data.ID,
		Data:     data.Data,
		Old:      data.Old,
		Metadata: data.Metadata,
	}, nil
}
