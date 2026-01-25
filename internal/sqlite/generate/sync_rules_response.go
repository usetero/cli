package main

import "sort"

// SyncRulesResponse is the response from /api/sync-rules/v1/current.
type SyncRulesResponse struct {
	Data struct {
		Current struct {
			BucketDefinitions []BucketDefinition `json:"bucket_definitions"`
		} `json:"current"`
	} `json:"data"`
}

// BucketDefinition represents a bucket in sync rules.
type BucketDefinition struct {
	DataQueries []DataQuery `json:"data_queries"`
}

// DataQuery represents a data query in sync rules.
type DataQuery struct {
	Table   QueryTable `json:"table"`
	Columns []string   `json:"columns"`
}

// QueryTable represents the table info in a data query.
type QueryTable struct {
	TablePattern string `json:"tablePattern"`
}

// Tables extracts the synced tables with their columns and types.
func (r *SyncRulesResponse) Tables(columnTypes map[string]map[string]ColumnType) []Table {
	// Collect unique tables and their columns
	tableColumns := make(map[string][]string)

	for _, bucket := range r.Data.Current.BucketDefinitions {
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

	var tables []Table
	for tableName, columns := range tableColumns {
		table := Table{
			Name:       tableName,
			StructName: toStructName(tableName),
			FileName:   toFileName(tableName),
		}

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

		for _, colName := range columns {
			colType := columnTypes[tableName][colName]
			col := Column{
				Name:       colName,
				FieldName:  toFieldName(colName),
				GoType:     toGoType(colType.SQLiteType, colType.PgType),
				SQLiteType: toSQLiteType(colType.SQLiteType),
				PgType:     colType.PgType,
			}
			table.Columns = append(table.Columns, col)
		}

		tables = append(tables, table)
	}

	return tables
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
