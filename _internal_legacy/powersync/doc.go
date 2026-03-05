// Package powersync provides background sync using PowerSync.
//
// It takes a sqlite.DB, loads the PowerSync extension, and keeps the
// database in sync with the server via HTTP streaming.
//
// You don't query through powersync - just use sqlite directly.
//
// Internal architecture:
//   - syncer.go: public API, lifecycle wiring, dependencies
//   - syncer_run.go: retry/backoff loop and token refresh paths
//   - syncer_stream.go: session and stream processing
//   - syncer_instructions.go: extension instruction application
//   - syncer_controlplane.go: serialized extension control-plane calls
//   - stream_capture.go: optional raw NDJSON stream capture utilities
//
// Optional fixture capture:
//   - Use `tero internal powersync capture` to record raw stream fixtures.
//   - Use `task internal:powersync:capture` as a wrapper in dev/prd.
//
// Subpackages:
//   - api: HTTP client for PowerSync service
//   - extension: SQLite extension interface and wire types
//   - db: Local database operations (CRUD queue, batch completion)
//
//go:generate go run ./extension/generate
package powersync
