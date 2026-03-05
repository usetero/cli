package db

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
)

// CompleteBatch finalizes a CRUD upload by atomically:
// 1. Deleting uploaded entries from ps_crud
// 2. Setting target_op to the write checkpoint
//
// This must be called after uploading entries and fetching a write checkpoint
// from the server. The checkpoint signals to sync that uploads are complete.
func CompleteBatch(ctx context.Context, db sqlite.DB, lastEntryID int64, checkpoint string) error {
	return db.WithTx(ctx, func(tx *sqlite.Tx) error {
		// Delete all uploaded entries
		if _, err := tx.Exec(ctx, "DELETE FROM ps_crud WHERE id <= ?", lastEntryID); err != nil {
			return fmt.Errorf("delete crud entries: %w", err)
		}

		// Set target_op to the checkpoint
		// This tells sync_local that we're waiting for this checkpoint from the server
		if _, err := tx.Exec(ctx,
			"UPDATE ps_buckets SET target_op = CAST(? AS INTEGER) WHERE name = '$local'",
			checkpoint,
		); err != nil {
			return fmt.Errorf("update target_op: %w", err)
		}

		return nil
	})
}

// GetClientID returns the PowerSync client ID for this database.
// This ID is used when fetching write checkpoints from the server.
func GetClientID(ctx context.Context, db sqlite.DB) (string, error) {
	var clientID string
	err := db.QueryRow(ctx, "SELECT powersync_client_id()").Scan(&clientID)
	if err != nil {
		return "", fmt.Errorf("get client id: %w", err)
	}
	return clientID, nil
}
