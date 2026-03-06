package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type crudRow struct {
	ID   int64
	TxID *int64
	Data string
}

type crudPayload struct {
	Op       string         `json:"op"`
	Type     string         `json:"type"`
	ID       string         `json:"id"`
	Data     map[string]any `json:"data,omitempty"`
	Old      map[string]any `json:"old,omitempty"`
	Metadata *string        `json:"metadata,omitempty"`
}

// Store provides access to PowerSync local tables.
type Store struct {
	db *sqlite.DB
}

// NewStore builds a PowerSync DB store on top of an initialized SQLite DB.
func NewStore(db *sqlite.DB) *Store {
	if db == nil {
		panic("powersync db store requires db")
	}
	return &Store{db: db}
}

// HasPendingMutations reports whether ps_crud has queued rows.
func (s *Store) HasPendingMutations(ctx context.Context) (bool, error) {
	var count int64
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM ps_crud").Scan(&count); err != nil {
		return false, wrap("count pending", err)
	}
	return count > 0, nil
}

// NextMutation returns the first queued mutation, or nil when the queue is empty.
func (s *Store) NextMutation(ctx context.Context) (*Mutation, error) {
	var row crudRow
	err := s.db.QueryRow(ctx, "SELECT id, tx_id, data FROM ps_crud ORDER BY id LIMIT 1").Scan(&row.ID, &row.TxID, &row.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, wrap("select next mutation", err)
	}
	mutation, err := parseMutation(row)
	if err != nil {
		return nil, wrap("parse next mutation", err)
	}
	return mutation, nil
}

// NextMutationBatch returns one logical upload batch.
// If the first row has no tx_id, the batch has a single mutation.
func (s *Store) NextMutationBatch(ctx context.Context) ([]Mutation, error) {
	first, err := s.NextMutation(ctx)
	if err != nil || first == nil {
		return nil, err
	}
	if first.TxID == nil {
		return []Mutation{*first}, nil
	}

	rows, err := s.db.Query(ctx, "SELECT id, tx_id, data FROM ps_crud WHERE tx_id = ? ORDER BY id", int64(*first.TxID))
	if err != nil {
		return nil, wrap("select batch by tx", err)
	}
	defer rows.Close()

	batch := make([]Mutation, 0)
	for rows.Next() {
		var row crudRow
		if err := rows.Scan(&row.ID, &row.TxID, &row.Data); err != nil {
			return nil, wrap("scan batch row", err)
		}
		mutation, err := parseMutation(row)
		if err != nil {
			return nil, wrap("parse batch row", err)
		}
		batch = append(batch, *mutation)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate batch rows", err)
	}
	return batch, nil
}

// PendingMutations returns all queued mutations ordered by id.
func (s *Store) PendingMutations(ctx context.Context) ([]Mutation, error) {
	rows, err := s.db.Query(ctx, "SELECT id, tx_id, data FROM ps_crud ORDER BY id")
	if err != nil {
		return nil, wrap("select pending", err)
	}
	defer rows.Close()

	out := make([]Mutation, 0)
	for rows.Next() {
		var row crudRow
		if err := rows.Scan(&row.ID, &row.TxID, &row.Data); err != nil {
			return nil, wrap("scan pending row", err)
		}
		mutation, err := parseMutation(row)
		if err != nil {
			return nil, wrap("parse pending row", err)
		}
		out = append(out, *mutation)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate pending rows", err)
	}
	return out, nil
}

// CompleteUploadedBatch atomically deletes uploaded rows and sets $local target_op.
func (s *Store) CompleteUploadedBatch(ctx context.Context, lastMutationID MutationID, checkpoint OpID) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		var currentTargetOp int64
		if err := tx.QueryRowContext(ctx, "SELECT target_op FROM ps_buckets WHERE name = ?", string(LocalBucket)).Scan(&currentTargetOp); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("local bucket is required before completing uploaded batch")
			}
			return wrap("read local target_op", err)
		}
		if int64(checkpoint) < currentTargetOp {
			return fmt.Errorf("checkpoint regression: current target_op=%d next=%d", currentTargetOp, checkpoint)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM ps_crud WHERE id <= ?", int64(lastMutationID)); err != nil {
			return wrap("delete uploaded batch", err)
		}
		result, err := tx.ExecContext(ctx,
			"UPDATE ps_buckets SET target_op = ? WHERE name = ?",
			int64(checkpoint),
			string(LocalBucket),
		)
		if err != nil {
			return wrap("set local target_op", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return wrap("read local target_op rows affected", err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("expected exactly one local bucket row, updated %d", rowsAffected)
		}
		return nil
	})
}

// ClientID returns powersync_client_id() for this database.
func (s *Store) ClientID(ctx context.Context) (string, error) {
	var id string
	if err := s.db.QueryRow(ctx, "SELECT powersync_client_id()").Scan(&id); err != nil {
		return "", wrap("get client id", err)
	}
	return id, nil
}

// CheckHealth validates critical PowerSync local invariants.
func (s *Store) CheckHealth(ctx context.Context) error {
	var txCount int
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM ps_tx WHERE id = 1").Scan(&txCount); err != nil {
		return wrap("check ps_tx", err)
	}
	if txCount == 0 {
		return fmt.Errorf("%w: ps_tx missing required row", ErrCorrupt)
	}

	hasPending, err := s.HasPendingMutations(ctx)
	if err != nil {
		return err
	}
	if hasPending {
		return nil
	}

	var targetOp, lastOp int64
	err = s.db.QueryRow(ctx, "SELECT target_op, last_op FROM ps_buckets WHERE name = ?", string(LocalBucket)).Scan(&targetOp, &lastOp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return wrap("check local bucket", err)
	}
	if targetOp > lastOp {
		return fmt.Errorf("%w: %s bucket is stuck (target_op=%d > last_op=%d)", ErrCorrupt, LocalBucket, targetOp, lastOp)
	}
	return nil
}

func parseMutation(row crudRow) (*Mutation, error) {
	var payload crudPayload
	if err := json.Unmarshal([]byte(row.Data), &payload); err != nil {
		return nil, err
	}
	if payload.ID == "" {
		return nil, fmt.Errorf("mutation row id is required")
	}
	if payload.Type == "" {
		return nil, fmt.Errorf("mutation table type is required")
	}
	switch Operation(payload.Op) {
	case OperationPut, OperationPatch, OperationDelete:
	default:
		return nil, fmt.Errorf("invalid mutation op %q", payload.Op)
	}

	var txID *TransactionID
	if row.TxID != nil {
		v := TransactionID(*row.TxID)
		txID = &v
	}

	return &Mutation{
		ID:       MutationID(row.ID),
		TxID:     txID,
		Op:       Operation(payload.Op),
		Table:    TableName(payload.Type),
		RowID:    payload.ID,
		Data:     payload.Data,
		Old:      payload.Old,
		Metadata: payload.Metadata,
	}, nil
}
