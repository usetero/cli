package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// SQLite type constants from PowerSync schema API.
const (
	sqliteTypeText    = 2
	sqliteTypeInteger = 4
	sqliteTypeReal    = 8
)

// SchemaTable represents a table in the PowerSync schema.
type SchemaTable struct {
	Name    string         `json:"name"`
	Columns []SchemaColumn `json:"columns"`
	Indexes []SchemaIndex  `json:"indexes"`
}

// SchemaColumn represents a column in a PowerSync table.
type SchemaColumn struct {
	Name string `json:"name"`
	Type string `json:"type"` // "text", "integer", "real"
}

// SchemaIndex represents an index on a PowerSync table.
type SchemaIndex struct {
	Name    string              `json:"name"`
	Columns []SchemaIndexColumn `json:"columns"`
}

// SchemaIndexColumn represents a column in an index.
type SchemaIndexColumn struct {
	Name      string `json:"name"`
	Ascending bool   `json:"ascending"`
	Type      string `json:"type"`
}

// schemaResponse is the response from /api/admin/v1/schema.
type schemaResponse struct {
	Data struct {
		Connections []struct {
			Schemas []struct {
				Name   string `json:"name"`
				Tables []struct {
					Name    string `json:"name"`
					Columns []struct {
						Name       string `json:"name"`
						SQLiteType int    `json:"sqlite_type"`
						PgType     string `json:"pg_type"`
					} `json:"columns"`
				} `json:"tables"`
			} `json:"schemas"`
		} `json:"connections"`
	} `json:"data"`
}

// syncRulesResponse is the response from /api/sync-rules/v1/current.
type syncRulesResponse struct {
	Data struct {
		Current struct {
			BucketDefinitions []struct {
				DataQueries []struct {
					Table struct {
						TablePattern string `json:"tablePattern"`
					} `json:"table"`
					Columns []string `json:"columns"`
				} `json:"data_queries"`
			} `json:"bucket_definitions"`
		} `json:"current"`
	} `json:"data"`
}

// columnType holds type information for a column.
type columnType struct {
	sqliteType int
	pgType     string
}

// FetchSchemaJSON fetches the schema from the PowerSync service and returns it as JSON.
// The returned JSON is in the format expected by powersync_replace_schema().
func FetchSchemaJSON(ctx context.Context, endpoint, token string) (string, error) {
	// Fetch column types from schema API
	columnTypes, err := fetchColumnTypes(ctx, endpoint, token)
	if err != nil {
		return "", fmt.Errorf("fetch column types: %w", err)
	}

	// Fetch synced tables from sync rules API
	tables, err := fetchSyncedTables(ctx, endpoint, token, columnTypes)
	if err != nil {
		return "", fmt.Errorf("fetch synced tables: %w", err)
	}

	// Apply client-side indexes for query performance
	tables = applyClientIndexes(tables)

	// Build schema JSON
	schema := struct {
		Tables []SchemaTable `json:"tables"`
	}{
		Tables: tables,
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("marshal schema: %w", err)
	}

	return string(data), nil
}

// fetchColumnTypes fetches column type information from the schema API.
func fetchColumnTypes(ctx context.Context, endpoint, token string) (map[string]map[string]columnType, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api/admin/v1/schema", strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("schema API returned %d: %s", resp.StatusCode, body)
	}

	var result schemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Build column type map
	columnTypes := make(map[string]map[string]columnType)
	for _, conn := range result.Data.Connections {
		for _, schema := range conn.Schemas {
			for _, table := range schema.Tables {
				if columnTypes[table.Name] == nil {
					columnTypes[table.Name] = make(map[string]columnType)
				}
				for _, col := range table.Columns {
					columnTypes[table.Name][col.Name] = columnType{
						sqliteType: col.SQLiteType,
						pgType:     col.PgType,
					}
				}
			}
		}
	}

	return columnTypes, nil
}

// fetchSyncedTables fetches the synced tables from the sync rules API.
func fetchSyncedTables(ctx context.Context, endpoint, token string, columnTypes map[string]map[string]columnType) ([]SchemaTable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/sync-rules/v1/current", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sync-rules API returned %d: %s", resp.StatusCode, body)
	}

	var result syncRulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Collect unique tables and their columns
	tableColumns := make(map[string][]string)
	for _, bucket := range result.Data.Current.BucketDefinitions {
		for _, query := range bucket.DataQueries {
			tableName := query.Table.TablePattern
			existing := tableColumns[tableName]
			for _, col := range query.Columns {
				if !contains(existing, col) {
					existing = append(existing, col)
				}
			}
			tableColumns[tableName] = existing
		}
	}

	// Build schema tables
	var tables []SchemaTable
	for tableName, columns := range tableColumns {
		// Sort columns: id first, then alphabetically
		sort.Slice(columns, func(i, j int) bool {
			if columns[i] == "id" {
				return true
			}
			if columns[j] == "id" {
				return false
			}
			return columns[i] < columns[j]
		})

		var schemaColumns []SchemaColumn
		for _, colName := range columns {
			colType := columnTypes[tableName][colName]
			schemaColumns = append(schemaColumns, SchemaColumn{
				Name: colName,
				Type: toSQLiteTypeName(colType.sqliteType),
			})
		}

		tables = append(tables, SchemaTable{
			Name:    tableName,
			Columns: schemaColumns,
		})
	}

	// Sort tables by name for deterministic output
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return tables, nil
}

// toSQLiteTypeName converts a SQLite type constant to a string.
func toSQLiteTypeName(sqliteType int) string {
	switch sqliteType {
	case sqliteTypeInteger:
		return "integer"
	case sqliteTypeReal:
		return "real"
	default:
		return "text"
	}
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
