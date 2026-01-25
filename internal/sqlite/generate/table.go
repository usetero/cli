package main

// Table represents a synced table for code generation.
type Table struct {
	Name       string   // SQL table name (e.g., "log_events")
	StructName string   // Go struct name (e.g., "LogEvent")
	FileName   string   // Output file name (e.g., "log_events.go")
	Columns    []Column // Columns to generate
}

// NeedsJSON returns true if any column uses json.RawMessage.
func (t Table) NeedsJSON() bool {
	for _, col := range t.Columns {
		if col.GoType == "json.RawMessage" {
			return true
		}
	}
	return false
}

// Column represents a column for code generation.
type Column struct {
	Name       string // SQL column name (e.g., "account_id")
	FieldName  string // Go field name (e.g., "AccountID")
	GoType     string // Go type (e.g., "string", "int64")
	SQLiteType string // SQLite type (e.g., "text", "integer")
	PgType     string // Postgres type for comments (e.g., "uuid")
}
