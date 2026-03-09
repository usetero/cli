package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

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
	Type string `json:"type"`
}

// SchemaIndex represents an index on a PowerSync table.
type SchemaIndex struct {
	Name    string              `json:"name"`
	Columns []SchemaIndexColumn `json:"columns"`
}

// SchemaIndexColumn represents a column in a PowerSync index.
type SchemaIndexColumn struct {
	Name      string `json:"name"`
	Ascending bool   `json:"ascending"`
	Type      string `json:"type"`
}

type schemaResponse struct {
	Data struct {
		Connections []struct {
			Schemas []struct {
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

type columnType struct {
	sqliteType int
	pgType     string
}

// FetchSchemaJSON fetches the live PowerSync schema and returns the JSON payload
// expected by powersync_replace_schema().
func FetchSchemaJSON(ctx context.Context, endpoint string, token AccessToken) (string, error) {
	columnTypes, err := fetchColumnTypes(ctx, endpoint, token)
	if err != nil {
		return "", fmt.Errorf("fetch column types: %w", err)
	}

	tables, err := fetchSyncedTables(ctx, endpoint, token, columnTypes)
	if err != nil {
		return "", fmt.Errorf("fetch synced tables: %w", err)
	}

	schema := struct {
		Tables []SchemaTable `json:"tables"`
	}{
		Tables: applyClientIndexes(tables),
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("marshal schema: %w", err)
	}
	return string(data), nil
}

func fetchColumnTypes(ctx context.Context, endpoint string, token AccessToken) (map[string]map[string]columnType, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(endpoint, "/")+"/api/admin/v1/schema", strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+string(token))
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

func fetchSyncedTables(ctx context.Context, endpoint string, token AccessToken, columnTypes map[string]map[string]columnType) ([]SchemaTable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(endpoint, "/")+"/api/sync-rules/v1/current", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

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

	var tables []SchemaTable
	for tableName, columns := range tableColumns {
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

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return tables, nil
}

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

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
