package main

import "strings"

// SQLite type constants from PowerSync.
const (
	sqliteTypeText    = 2
	sqliteTypeInteger = 4
	sqliteTypeReal    = 8
)

// toStructName converts a table name to a Go struct name.
// Example: "log_events" -> "LogEvent"
func toStructName(tableName string) string {
	parts := strings.Split(tableName, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(part[1:])
		}
	}

	name := result.String()

	// Singularize common patterns
	if strings.HasSuffix(name, "ies") {
		name = name[:len(name)-3] + "y"
	} else if strings.HasSuffix(name, "ses") {
		name = name[:len(name)-2]
	} else if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		name = name[:len(name)-1]
	}

	return name
}

// toFieldName converts a column name to a Go field name.
// Example: "account_id" -> "AccountID"
func toFieldName(columnName string) string {
	parts := strings.Split(columnName, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			upper := strings.ToUpper(part)
			// Handle common abbreviations
			if upper == "ID" || upper == "URL" || upper == "API" || upper == "JWT" || upper == "AVG" {
				result.WriteString(upper)
			} else {
				result.WriteString(strings.ToUpper(part[:1]))
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}

// toGoType converts a SQLite type to a Go type.
func toGoType(sqliteType int, pgType string) string {
	// Special case for jsonb
	if pgType == "jsonb" {
		return "json.RawMessage"
	}

	switch sqliteType {
	case sqliteTypeInteger:
		return "int64"
	case sqliteTypeReal:
		return "float64"
	default:
		return "string"
	}
}

// toSQLiteType converts a SQLite type constant to a string.
func toSQLiteType(sqliteType int) string {
	switch sqliteType {
	case sqliteTypeInteger:
		return "integer"
	case sqliteTypeReal:
		return "real"
	default:
		return "text"
	}
}

// toFileName converts a table name to a singular generated file name.
// Example: "log_events" -> "log_event.generated.go"
func toFileName(tableName string) string {
	name := tableName

	// Singularize common patterns
	if strings.HasSuffix(name, "ies") {
		name = name[:len(name)-3] + "y"
	} else if strings.HasSuffix(name, "ses") {
		name = name[:len(name)-2]
	} else if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		name = name[:len(name)-1]
	}

	return name + ".generated.go"
}
