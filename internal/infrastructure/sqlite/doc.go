//go:generate go run ./generate

// Package sqlite provides the low-level SQLite kernel: open/close, local
// runtime migrations, transactions, timeouts, and storage paths.
//
// PowerSync-projected tables are not created here. They come from the embedded
// PowerSync schema artifact and are applied explicitly by the session/sync
// runtime after the database is opened. The generated schema.sql,
// jsonb_schema.json, and powersync_schema.json files are copied from the
// sibling control-plane repo during generation. schema.sql is used for sqlc and
// query-tool descriptions; jsonb_schema.json is the machine-readable payload
// contract for synced JSONB columns; powersync_schema.json is the exact
// client-side schema applied by the PowerSync extension. None of these files is
// replayed as a runtime migration.
package sqlite
