package main

import (
	"fmt"
	"regexp"
	"strings"
)

// annotateSchema adds comments to the SQLite schema.
// SQLite schema is the source of truth; comments are enrichment from Postgres.
func annotateSchema(schema string, comments map[string]string, sqliteVersion string) string {
	var result strings.Builder

	result.WriteString("-- Tero Local Catalog Schema\n")
	result.WriteString("-- Auto-generated with comments from control plane. DO NOT EDIT.\n")
	result.WriteString("--\n")
	result.WriteString(fmt.Sprintf("-- SQLite %s with full support for: CTEs, window functions, JSON functions,\n", sqliteVersion))
	result.WriteString("-- RETURNING, generated columns, and all modern SQL features.\n")
	result.WriteString("--\n")
	result.WriteString("-- This is a LOCAL SQLite database synced from the Tero control plane.\n")
	result.WriteString("-- All data is scoped to the authenticated user's account.\n")
	result.WriteString("--\n")
	result.WriteString("-- Key concepts:\n")
	result.WriteString("--   services      - Applications producing logs (e.g., 'checkout-service')\n")
	result.WriteString("--   log_events    - Distinct event patterns within a service\n")
	result.WriteString("--   policies      - AI-identified waste (health checks, duplicate fields, bloat)\n")
	result.WriteString("--   *_cache       - Pre-computed status and metrics (query these for current state)\n")
	result.WriteString("--\n")
	result.WriteString("-- IMPORTANT: All queries are READ-ONLY. This is a local sync of server data.\n")
	result.WriteString("\n")

	lines := strings.Split(schema, "\n")
	var currentTable string
	createTableRegex := regexp.MustCompile(`^CREATE TABLE (\w+)`)
	columnRegex := regexp.MustCompile(`^\s+(\w+)\s+`)

	for _, line := range lines {
		// Skip original header comments
		if strings.HasPrefix(line, "-- Code generated") || strings.HasPrefix(line, "-- Source:") {
			continue
		}

		// Track current table
		if match := createTableRegex.FindStringSubmatch(line); match != nil {
			currentTable = match[1]
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Annotate column if we have a comment
		if currentTable != "" {
			if match := columnRegex.FindStringSubmatch(line); match != nil {
				colName := match[1]
				key := currentTable + "." + colName

				// For cache tables, also try the base table name (without _cache suffix)
				comment, ok := comments[key]
				if !ok && strings.HasSuffix(currentTable, "_cache") {
					baseTable := strings.TrimSuffix(currentTable, "_cache")
					comment, ok = comments[baseTable+"."+colName]
				}

				if ok {
					trimmed := strings.TrimSuffix(line, ",")
					hasComma := strings.HasSuffix(line, ",")
					if hasComma {
						result.WriteString(fmt.Sprintf("%s, -- %s\n", trimmed, comment))
					} else {
						result.WriteString(fmt.Sprintf("%s -- %s\n", line, comment))
					}
					continue
				}
			}
		}

		// Reset on table end
		if strings.HasPrefix(line, ");") {
			currentTable = ""
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}
