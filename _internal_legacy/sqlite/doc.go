// Package sqlite provides the local SQLite database for the CLI.
//
// This is the primary data layer. PowerSync keeps it in sync with the server,
// but you query it directly like any SQLite database.
//
// # Code Generation
//
// The schema types are generated from the PowerSync service. To regenerate:
//
//	doppler run -- go generate ./internal/sqlite
//
// This fetches the current schema and sync rules from the PowerSync API
// and generates Go types for each synced table.
package sqlite

//go:generate go run ./generate
