package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
)

// LogExample is a parsed log record for rendering in cards and other views.
// Parsed from the log_events.examples JSON column.
type LogExample struct {
	Body       string         `json:"body"`
	Severity   string         `json:"severity_text"`
	Attributes map[string]any `json:"attributes"`
}

// Attr is a single flattened log attribute for display.
type Attr struct {
	Key   string
	Value string
}

// infrastructurePrefixes are top-level attribute key prefixes that describe
// infrastructure rather than application behavior. Matches the control plane's
// InfrastructureAttributeKeys list in otel/log_record_filter.go.
var infrastructurePrefixes = []string{
	"container", "host", "hostname", "k8s", "os", "cloud", "deployment",
	"process", "thread",
	"otel", "telemetry", "status", "agent", "collector", "log",
	"network", "source",
}

// httpHeaderKeys are exact attribute keys for HTTP transport headers that
// frequently leak into log attributes from request/response contexts.
// These are low-value for diagnostics — strip them alongside infrastructure.
var httpHeaderKeys = map[string]bool{
	"Accept":          true,
	"Content-Length":  true,
	"Content-Type":    true,
	"Authorization":   true,
	"User-Agent":      true,
	"Cache-Control":   true,
	"Connection":      true,
	"Accept-Encoding": true,
	"Accept-Language": true,
}

// ParseLogExamples parses the log_events.examples JSON column and returns
// clean examples with infrastructure/resource/scope attributes stripped.
func ParseLogExamples(jsonStr string) []LogExample {
	// Parse full log records to access all fields.
	var raw []struct {
		Body       string         `json:"body"`
		Severity   string         `json:"severity_text"`
		Attributes map[string]any `json:"attributes"`
		// Parsed but discarded — we only want message + log-level attributes.
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil
	}

	examples := make([]LogExample, 0, len(raw))
	for _, r := range raw {
		ex := LogExample{
			Body:       r.Body,
			Severity:   r.Severity,
			Attributes: stripInfrastructure(r.Attributes),
		}
		examples = append(examples, ex)
	}
	return examples
}

// IsEmpty returns true if the example has no body and no attributes.
func (e LogExample) IsEmpty() bool {
	return e.Body == "" && len(e.Attributes) == 0
}

// FlatAttrs returns all attributes flattened and ordered by relevance.
// Relevant keys appear first (in the order provided), then remaining
// attributes sorted alphabetically. Returns all attributes — callers
// decide how many to display.
func (e LogExample) FlatAttrs(relevantKeys []FieldPath) []Attr {
	if len(e.Attributes) == 0 {
		return nil
	}

	flat := flattenMap("", e.Attributes)
	ordered := orderKeys(flat, relevantKeys)

	attrs := make([]Attr, len(ordered))
	for i, k := range ordered {
		attrs[i] = Attr{Key: k, Value: flat[k]}
	}
	return attrs
}

// orderKeys returns attribute keys with relevantKeys first (preserving order,
// skipping any not present in flat), then remaining keys alphabetically.
func orderKeys(flat map[string]string, relevantKeys []FieldPath) []string {
	if len(relevantKeys) == 0 {
		keys := make([]string, 0, len(flat))
		for k := range flat {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}

	seen := make(map[string]bool, len(relevantKeys))
	ordered := make([]string, 0, len(flat))

	// Relevant keys first, in order provided.
	for _, rk := range relevantKeys {
		key := rk.Key()
		if _, ok := flat[key]; ok && !seen[key] {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}

	// Remaining keys alphabetically.
	rest := make([]string, 0, len(flat)-len(ordered))
	for k := range flat {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)

	return ordered
}

// stripInfrastructure removes infrastructure attribute keys by prefix.
func stripInfrastructure(attrs map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		if isInfrastructureKey(k) {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// isInfrastructureKey checks if a key matches any infrastructure prefix,
// known HTTP header, or X-prefixed transport header.
func isInfrastructureKey(key string) bool {
	if httpHeaderKeys[key] {
		return true
	}
	if strings.HasPrefix(key, "X-") {
		return true
	}
	for _, prefix := range infrastructurePrefixes {
		if key == prefix || strings.HasPrefix(key, prefix+".") {
			return true
		}
	}
	return false
}

// flattenMap recursively flattens nested maps with dot-separated keys.
func flattenMap(prefix string, m map[string]any) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]any); ok {
			maps.Copy(out, flattenMap(key, nested))
		} else {
			out[key] = formatValue(v)
		}
	}
	return out
}

// formatValue renders an attribute value as a display string.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}
