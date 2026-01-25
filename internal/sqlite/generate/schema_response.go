package main

// SchemaResponse is the response from /api/admin/v1/schema.
type SchemaResponse struct {
	Data struct {
		Connections []Connection `json:"connections"`
	} `json:"data"`
}

// Connection represents a database connection in the schema response.
type Connection struct {
	Schemas []DatabaseSchema `json:"schemas"`
}

// DatabaseSchema represents a database schema (e.g., "public").
type DatabaseSchema struct {
	Name   string          `json:"name"`
	Tables []DatabaseTable `json:"tables"`
}

// DatabaseTable represents a table in the database schema.
type DatabaseTable struct {
	Name    string           `json:"name"`
	Columns []DatabaseColumn `json:"columns"`
}

// DatabaseColumn represents a column in the database schema.
type DatabaseColumn struct {
	Name         string `json:"name"`
	SQLiteType   int    `json:"sqlite_type"`
	InternalType string `json:"internal_type"`
	PgType       string `json:"pg_type"`
}

// ColumnTypes returns a map of table -> column -> type info.
func (r *SchemaResponse) ColumnTypes() map[string]map[string]ColumnType {
	result := make(map[string]map[string]ColumnType)

	for _, conn := range r.Data.Connections {
		for _, schema := range conn.Schemas {
			for _, table := range schema.Tables {
				if result[table.Name] == nil {
					result[table.Name] = make(map[string]ColumnType)
				}
				for _, col := range table.Columns {
					result[table.Name][col.Name] = ColumnType{
						SQLiteType: col.SQLiteType,
						PgType:     col.PgType,
					}
				}
			}
		}
	}

	return result
}

// ColumnType holds type information for a column.
type ColumnType struct {
	SQLiteType int
	PgType     string
}
