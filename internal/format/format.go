// Package format provides shared formatting helpers for display values.
package format

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// Cost formats a dollar amount: $142, $9.4k, $1.2M.
func Cost(dollars float64) string {
	abs := math.Abs(dollars)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("$%.1fM", dollars/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("$%.1fk", dollars/1_000)
	default:
		return fmt.Sprintf("$%.0f", dollars)
	}
}

// YearlyCost formats an hourly USD rate as yearly: $0/yr, $1.7k/yr, $110k/yr.
func YearlyCost(costPerHour float64) string {
	yearly := costPerHour * 8760
	if yearly < 1 {
		return "$0/yr"
	}
	return Cost(yearly) + "/yr"
}

// YearlyCostPtr formats a nullable hourly rate: nil → "—", 0 → "$0/yr",
// positive → "~$1.7k/yr". Use this at display boundaries where the value
// may not yet be estimated (nil) vs estimated at zero.
func YearlyCostPtr(costPerHour *float64) string {
	if costPerHour == nil {
		return "—"
	}
	if *costPerHour <= 0 {
		return "$0/yr"
	}
	return "~" + YearlyCost(*costPerHour)
}

// Volume formats events/hr: 892, 45.3k, 2.1M.
func Volume(eventsPerHour float64) string {
	abs := math.Abs(eventsPerHour)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", eventsPerHour/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", eventsPerHour/1_000)
	default:
		return fmt.Sprintf("%.0f", eventsPerHour)
	}
}

// Bytes formats byte amounts: 540 B, 12.4 KB, 1.2 MB, 3.4 GB.
func Bytes(b float64) string {
	abs := math.Abs(b)
	switch {
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", b/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1f MB", b/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1f KB", b/1_000)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// Count renders an integer, showing "—" for zero.
func Count(n int64) string {
	if n == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", n)
}

// TitleCase converts a snake_case slug to Title Case.
// "bot_traffic" → "Bot Traffic", "pii_leakage" → "Pii Leakage".
func TitleCase(slug string) string {
	words := strings.Split(slug, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = string(unicode.ToUpper(rune(w[0]))) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
