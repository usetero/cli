//go:generate go run ./generate

// Package sqlite provides the low-level SQLite kernel: open/close, local
// runtime migrations, transactions, timeouts, and storage paths.
//
// PowerSync-projected tables are not created here. They come from the embedded
// PowerSync schema and are applied explicitly by the session/sync runtime after
// the database is opened. The reflected schema.sql file is generated for sqlc
// and query-tool descriptions, not replayed as a runtime migration.
package sqlite
