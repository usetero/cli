package domain

import (
	"encoding/json"
	"strings"
)

// FieldPath is the canonical representation of an attribute path in a log event.
// It wraps the segment slice that the control plane stores (e.g. ["attributes", "http", "status"])
// and provides a stable dot-key for display and matching (e.g. "http.status").
//
// Three representations exist in the wild:
//
//  1. JSON array of strings — analysis JSON fields (["attributes", "http", "status"])
//  2. PostgreSQL text array literal — log_event_fields.field_path ("{attributes,http,status}")
//  3. Dot-key — display and matching ("http.status")
//
// FieldPath normalizes all three into one type. Construct via:
//   - JSON unmarshal (handles #1 automatically)
//   - ParseFieldPathPg (handles #2)
//   - NewFieldPath (wraps a raw []string)
type FieldPath []string

// NewFieldPath wraps a segment slice into a FieldPath.
func NewFieldPath(segments []string) FieldPath {
	return FieldPath(segments)
}

// Key returns the dot-separated display key, stripping the leading "attributes"
// prefix since log examples already flatten under that namespace.
//
//	["attributes", "http", "status"] → "http.status"
//	["http", "status"]              → "http.status"
//	["body"]                        → "body"
func (p FieldPath) Key() string {
	segments := []string(p)
	if len(segments) == 0 {
		return ""
	}
	if segments[0] == "attributes" && len(segments) > 1 {
		segments = segments[1:]
	}
	return strings.Join(segments, ".")
}

// String implements fmt.Stringer via Key().
func (p FieldPath) String() string { return p.Key() }

// Segments returns the underlying path segments.
func (p FieldPath) Segments() []string { return []string(p) }

// IsEmpty returns true if the path has no segments.
func (p FieldPath) IsEmpty() bool { return len(p) == 0 }

// ParseFieldPathPg parses a PostgreSQL text array literal into a FieldPath.
//
//	"{attributes,http,status}" → ["attributes", "http", "status"]
//	""                         → nil
func ParseFieldPathPg(s string) FieldPath {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return nil
	}
	// Strip surrounding braces.
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	parts := strings.Split(s, ",")
	segments := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			segments = append(segments, p)
		}
	}
	if len(segments) == 0 {
		return nil
	}
	return FieldPath(segments)
}

// UnmarshalJSON supports direct deserialization from JSON arrays (["attributes", "http", "status"]).
// This means analysis structs can use FieldPath directly instead of []string.
func (p *FieldPath) UnmarshalJSON(data []byte) error {
	var segments []string
	if err := json.Unmarshal(data, &segments); err != nil {
		return err
	}
	*p = segments
	return nil
}

// MarshalJSON serializes as a JSON array of strings.
func (p FieldPath) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(p))
}
