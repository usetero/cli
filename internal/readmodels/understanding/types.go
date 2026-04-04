package understanding

import "github.com/usetero/cli/internal/domains/catalog"

type SnapshotRequest struct {
	ServiceOffset int
	ServiceLimit  int
	EventOffset   int
	EventLimit    int
}

type Snapshot struct {
	TotalServices int
	TotalEvents   int
	Services      []ServiceBand
}

type ServiceBand struct {
	ID         catalog.ServiceID
	Name       string
	EventCount int
	Events     []EventGlyph
}

type GlyphKind uint8

const (
	GlyphNormal GlyphKind = iota
	GlyphElevated
	GlyphFlagged
)

type EventGlyph struct {
	ID       catalog.LogEventID
	Name     string
	Severity string
	Kind     GlyphKind
}

type EventDetail struct {
	Event catalog.LogEvent
}
