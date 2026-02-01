package powersync

import (
	"bytes"
	"encoding/json"
)

// InstructionType constants for Instruction.Type.
const (
	InstructionEstablishSyncStream = "EstablishSyncStream"
	InstructionFetchCredentials    = "FetchCredentials"
	InstructionCloseSyncStream     = "CloseSyncStream"
	InstructionFlushFileSystem     = "FlushFileSystem"
	InstructionDidCompleteSync     = "DidCompleteSync"
	InstructionUpdateSyncStatus    = "UpdateSyncStatus"
	InstructionLogLine             = "LogLine"
)

// Instruction is a command returned by powersync_control.
// The extension returns instructions as tagged enums: {"InstructionType": {fields...}}
type Instruction struct {
	Type string
	// Fields vary by type
	Request        *StreamingSyncRequest
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
		i.Type = typ

		// Parse the payload based on type using strict unmarshaling
		switch typ {
		case InstructionEstablishSyncStream:
			var p struct {
				Request *StreamingSyncRequest `json:"request"`
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
