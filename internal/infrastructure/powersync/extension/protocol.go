package extension

import (
	"bytes"
	"encoding/json"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
)

type InstructionType string

const (
	InstructionEstablishSyncStream InstructionType = "EstablishSyncStream"
	InstructionFetchCredentials    InstructionType = "FetchCredentials"
	InstructionCloseSyncStream     InstructionType = "CloseSyncStream"
	InstructionFlushFileSystem     InstructionType = "FlushFileSystem"
	InstructionDidCompleteSync     InstructionType = "DidCompleteSync"
	InstructionUpdateSyncStatus    InstructionType = "UpdateSyncStatus"
	InstructionLogLine             InstructionType = "LogLine"
)

// LogSeverity is the extension log line severity.
type LogSeverity string

const (
	LogSeverityDebug LogSeverity = "debug"
	LogSeverityInfo  LogSeverity = "info"
	LogSeverityWarn  LogSeverity = "warn"
	LogSeverityError LogSeverity = "error"
)

// Instruction is a command returned by powersync_control.
type Instruction struct {
	Type           InstructionType
	Request        *psclient.SyncStreamRequest
	DidExpire      *bool
	HideDisconnect *bool
	SyncStatus     *SyncStatus
	Severity       LogSeverity
	Line           string
}

func unmarshalStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (i *Instruction) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for typ, payload := range raw {
		i.Type = InstructionType(typ)
		switch i.Type {
		case InstructionEstablishSyncStream:
			var p struct {
				Request *psclient.SyncStreamRequest `json:"request"`
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
			var wrapper struct {
				Status SyncStatus `json:"status"`
			}
			if err := unmarshalStrict(payload, &wrapper); err != nil {
				return err
			}
			i.SyncStatus = &wrapper.Status
		case InstructionLogLine:
			var p struct {
				Severity LogSeverity `json:"severity"`
				Line     string      `json:"line"`
			}
			if err := unmarshalStrict(payload, &p); err != nil {
				return err
			}
			i.Severity = p.Severity
			i.Line = p.Line
		}
		break
	}

	return nil
}

type SyncStatus struct {
	Connected      bool              `json:"connected"`
	Connecting     bool              `json:"connecting"`
	PriorityStatus []PriorityStatus  `json:"priority_status"`
	Downloading    *DownloadProgress `json:"downloading"`
	Streams        []StreamStatus    `json:"streams"`
}

type PriorityStatus struct {
	Priority     int   `json:"priority"`
	LastSyncedAt *int  `json:"last_synced_at"`
	HasSynced    *bool `json:"has_synced"`
}

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

type StreamProgress struct {
	Total      int `json:"total"`
	Downloaded int `json:"downloaded"`
}

type DownloadProgress struct {
	Buckets map[string]BucketProgress `json:"buckets"`
}

type BucketProgress struct {
	Priority    int `json:"priority"`
	AtLast      int `json:"at_last"`
	SinceLast   int `json:"since_last"`
	TargetCount int `json:"target_count"`
}

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
