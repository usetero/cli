//go:generate go run ./generate

// Package sqlite provides the low-level SQLite kernel: open/close, local
// runtime migrations, transactions, timeouts, and storage paths.
//
// PowerSync-projected tables are not created here. They come from the embedded
// PowerSync schema and are applied explicitly by the session/sync runtime after
// the database is opened. The authored schema.sql and jsonb_schema.json files
// are copied from the sibling control-plane repo during generation. schema.sql
// is used for sqlc and query-tool descriptions; jsonb_schema.json is the
// machine-readable payload contract for synced JSONB columns. Neither is
// replayed as a runtime migration.
package sqlite
