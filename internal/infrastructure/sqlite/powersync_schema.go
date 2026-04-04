package sqlite

import _ "embed"

// PowerSyncSchemaJSON contains the generated PowerSync client schema artifact
// applied by the local extension runtime.
//
// Generated via `task generate:schema` from the sibling control-plane repo.
//
//go:embed powersync_schema.json
var powerSyncSchemaJSON string

// PowerSyncSchemaJSON returns the generated PowerSync client schema artifact.
func PowerSyncSchemaJSON() string {
	return powerSyncSchemaJSON
}
