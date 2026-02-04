package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// fetchComments queries column comments from the control plane Postgres database.
// Returns a map of "table.column" -> "comment".
func fetchComments(dbURL string) (map[string]string, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer db.Close()

	// Query column comments from Postgres system catalogs
	rows, err := db.QueryContext(context.Background(), `
		SELECT
			c.relname AS table_name,
			a.attname AS column_name,
			d.description AS comment
		FROM pg_class c
		JOIN pg_attribute a ON a.attrelid = c.oid
		JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = a.attnum
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
		  AND c.relkind IN ('r', 'v')  -- tables and views
		  AND a.attnum > 0
		  AND d.description IS NOT NULL
		ORDER BY c.relname, a.attnum
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	comments := make(map[string]string)
	for rows.Next() {
		var table, column, comment string
		if err := rows.Scan(&table, &column, &comment); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		key := table + "." + column
		comments[key] = comment
	}

	return comments, rows.Err()
}
