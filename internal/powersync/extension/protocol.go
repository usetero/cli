// Package extension provides the interface to the PowerSync SQLite extension.
package extension

import (
	"bytes"
	"encoding/json"

	"github.com/usetero/cli/internal/powersync/api"
)

// InstructionType represents a PowerSync instruction type.
type InstructionType string

// InstructionType constants for Instruction.Type.
const (
	InstructionEstablishSyncStream InstructionType = "EstablishSyncStream"
	InstructionFetchCredentials    InstructionType = "FetchCredentials"
	InstructionCloseSyncStream     InstructionType = "CloseSyncStream"
	InstructionFlushFileSystem     InstructionType = "FlushFileSystem"
	InstructionDidCompleteSync     InstructionType = "DidCompleteSync"
	InstructionUpdateSyncStatus    InstructionType = "UpdateSyncStatus"
	InstructionLogLine             InstructionType = "LogLine"
)

// Instruction is a command returned by powersync_control.
// The extension returns instructions as tagged enums: {"InstructionType": {fields...}}
type Instruction struct {
	Type InstructionType
	// Fields vary by type
	Request        *api.SyncStreamRequest
	DidExpire      *bool
	HideDisconnect *bool
	SyncStatus     *SyncStatus
	Severity       string
	Line           string
}

// unmarshalStrict unmarshals JSON with DisallowUnknownFields to catch payload mismatches.
func unmarshalStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// UnmarshalJSON handles the serde-style tagged enum format from the extension.
// Example: {"EstablishSyncStream": {"request": {...}}}
func (i *Instruction) UnmarshalJSON(data []byte) error {
	// Parse as map to get the variant name (first key)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Should have exactly one key - the instruction type
	for typ, payload := range raw {
		i.Type = InstructionType(typ)

		// Parse the payload based on type using strict unmarshaling
		switch i.Type {
		case InstructionEstablishSyncStream:
			var p struct {
				Request *api.SyncStreamRequest `json:"request"`
			}
			if err := unmarshalStrict(payload, &p); err != nil {
				return err
			}
			i.Request = p.Request

		case InstructionFetchCredentials:
			var p struct {
				DidExpire *bool `json:"did_expire"`
			}
			if err := unmarshalStrict(payload, &p); err != nil {
				return err
			}
			i.DidExpire = p.DidExpire

		case InstructionCloseSyncStream:
			var p struct {
				HideDisconnect *bool `json:"hide_disconnect"`
			}
			if err := unmarshalStrict(payload, &p); err != nil {
				return err
			}
			i.HideDisconnect = p.HideDisconnect

		case InstructionUpdateSyncStatus:
			// Payload is wrapped: {"status": {...}}
			var wrapper struct {
				Status SyncStatus `json:"status"`
			}
			if err := unmarshalStrict(payload, &wrapper); err != nil {
				return err
			}
			i.SyncStatus = &wrapper.Status

		case InstructionLogLine:
			var p struct {
				Severity string `json:"severity"`
				Line     string `json:"line"`
			}
			if err := unmarshalStrict(payload, &p); err != nil {
				return err
			}
			i.Severity = p.Severity
			i.Line = p.Line

		default:
			// Unknown type - store raw for debugging but don't fail
		}

		// Only process the first key
		break
	}

	return nil
}

// SyncStatus represents the detailed sync state from UpdateSyncStatus instructions.
type SyncStatus struct {
	Connected      bool              `json:"connected"`
	Connecting     bool              `json:"connecting"`
	PriorityStatus []PriorityStatus  `json:"priority_status"`
	Downloading    *DownloadProgress `json:"downloading"`
	Streams        []StreamStatus    `json:"streams"`
}

// PriorityStatus represents sync status for a specific priority level.
type PriorityStatus struct {
	Priority     int   `json:"priority"`
	LastSyncedAt *int  `json:"last_synced_at"`
	HasSynced    *bool `json:"has_synced"`
}

// StreamStatus represents the status of a sync stream subscription.
type StreamStatus struct {
	Name                    string          `json:"name"`
	Parameters              *string         `json:"parameters"`
	Priority                int             `json:"priority"`
	Active                  bool            `json:"active"`
	IsDefault               bool            `json:"is_default"`
	HasExplicitSubscription bool            `json:"has_explicit_subscription"`
	ExpiresAt               *int            `json:"expires_at"`
	LastSyncedAt            *int            `json:"last_synced_at"`
	Progress                *StreamProgress `json:"progress"`
}

// StreamProgress represents download progress for a stream.
type StreamProgress struct {
	Total      int `json:"total"`
	Downloaded int `json:"downloaded"`
}

// DownloadProgress represents the current download progress.
type DownloadProgress struct {
	Buckets map[string]BucketProgress `json:"buckets"`
}

// BucketProgress represents progress for a single bucket.
type BucketProgress struct {
	Priority    int `json:"priority"`
	AtLast      int `json:"at_last"`
	SinceLast   int `json:"since_last"`
	TargetCount int `json:"target_count"`
}

// TotalProgress returns the total progress across all buckets.
// Returns (downloaded, total) counts.
func (d *DownloadProgress) TotalProgress() (int, int) {
	if d == nil {
		return 0, 0
	}
	var downloaded, total int
	for _, b := range d.Buckets {
		downloaded += b.SinceLast
		total += b.TargetCount
	}
	return downloaded, total
}
