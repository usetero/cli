// Package powersync provides background sync using PowerSync.
//
// It takes a sqlite.DB, loads the PowerSync extension, and keeps the
// database in sync with the server via HTTP streaming.
//
// You don't query through powersync - just use sqlite directly.
//
// Subpackages:
//   - api: HTTP client for PowerSync service
//   - extension: SQLite extension interface and wire types
//   - db: Local database operations (CRUD queue, batch completion)
//
//go:generate go run ./extension/generate
package powersync
