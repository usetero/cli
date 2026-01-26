// Package powersync provides background sync using PowerSync.
//
// It takes a sqlite.DB, loads the PowerSync extension, and keeps the
// database in sync with the server via HTTP streaming.
//
// You don't query through powersync - just use sqlite directly.
//
//go:generate go run ./generate
package powersync
