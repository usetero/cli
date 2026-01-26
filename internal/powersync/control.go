package powersync

import (
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
)

// Status represents the current state of the sync.
type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusSyncing      Status = "syncing"
	StatusReconnecting Status = "reconnecting"
	StatusError        Status = "error"
)

// ControlOp represents an operation for powersync_control.
type ControlOp string

const (
	// OpStart starts a sync stream. Payload: StartRequest (JSON).
	OpStart ControlOp = "start"
	// OpStop stops the current sync stream. Payload: none.
	OpStop ControlOp = "stop"
	// OpLineText forwards a JSON line from the sync service. Payload: string.
	OpLineText ControlOp = "line_text"
	// OpLineBinary forwards a BSON line from the sync service. Payload: []byte.
	OpLineBinary ControlOp = "line_binary"
	// OpRefreshedToken notifies that the auth token was refreshed. Payload: none.
	OpRefreshedToken ControlOp = "refreshed_token"
	// OpCompletedUpload notifies that CRUD upload completed. Payload: none.
	OpCompletedUpload ControlOp = "completed_upload"
	// OpUpdateSubscriptions updates stream subscriptions. Payload: JSON array.
	OpUpdateSubscriptions ControlOp = "update_subscriptions"
)

// ConnectionEvent represents a connection state change.
type ConnectionEvent string

const (
	// ConnectionEstablished indicates the sync stream connection was established.
	ConnectionEstablished ConnectionEvent = "established"
	// ConnectionEnded indicates the sync stream connection ended.
	ConnectionEnded ConnectionEvent = "end"
)

// StartRequest is the payload for OpStart.
type StartRequest struct {
	// Parameters are bucket parameters for the sync request.
	Parameters map[string]any `json:"parameters,omitempty"`
	// Schema defines the tables to sync.
	Schema json.RawMessage `json:"schema,omitempty"`
	// IncludeDefaults whether to request default streams.
	IncludeDefaults bool `json:"include_defaults"`
	// ActiveStreams are currently active stream subscriptions.
	ActiveStreams []StreamKey `json:"active_streams,omitempty"`
}

// StreamKey identifies a stream subscription.
type StreamKey struct {
	Name       string `json:"name"`
	Parameters string `json:"parameters,omitempty"`
}

// Instruction is a command returned by powersync_control.
// The extension returns instructions as tagged enums: {"InstructionType": {fields...}}
type Instruction struct {
	Type string
	// Fields vary by type
	Request        *StreamingSyncRequest
	DidExpire      *bool
	HideDisconnect *bool
	Status         json.RawMessage
	Severity       string
	Line           string
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

		// Parse the payload based on type
		switch typ {
		case InstructionEstablishSyncStream:
			var p struct {
				Request *StreamingSyncRequest `json:"request"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
			i.Request = p.Request

		case InstructionFetchCredentials:
			var p struct {
				DidExpire *bool `json:"did_expire"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
			i.DidExpire = p.DidExpire

		case InstructionCloseSyncStream:
			var p struct {
				HideDisconnect *bool `json:"hide_disconnect"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
			i.HideDisconnect = p.HideDisconnect

		case InstructionUpdateSyncStatus:
			var p struct {
				Status json.RawMessage `json:"status"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
			i.Status = p.Status

		case InstructionLogLine:
			var p struct {
				Severity string `json:"severity"`
				Line     string `json:"line"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
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

// StreamingSyncRequest is the request body for the sync stream endpoint.
// This is returned by the extension in EstablishSyncStream instructions.
type StreamingSyncRequest struct {
	Buckets         []BucketRequest     `json:"buckets"`
	IncludeChecksum bool                `json:"include_checksum"`
	RawData         bool                `json:"raw_data"`
	BinaryData      bool                `json:"binary_data"`
	ClientID        string              `json:"client_id"`
	Parameters      map[string]any      `json:"parameters,omitempty"`
	Streams         *StreamSubscription `json:"streams,omitempty"`
	AppMetadata     json.RawMessage     `json:"app_metadata,omitempty"`
}

// BucketRequest specifies a bucket to sync and the last known checkpoint.
type BucketRequest struct {
	Name  string `json:"name"`
	After string `json:"after"`
}

// StreamSubscription defines stream subscription preferences.
type StreamSubscription struct {
	IncludeDefaults bool                          `json:"include_defaults"`
	Subscriptions   []RequestedStreamSubscription `json:"subscriptions"`
}

// RequestedStreamSubscription is a request to subscribe to a stream.
type RequestedStreamSubscription struct {
	Stream           string `json:"stream"`
	Parameters       string `json:"parameters,omitempty"`
	OverridePriority *int   `json:"override_priority,omitempty"`
}

// Controller wraps the powersync_control SQLite function with type safety.
type Controller struct {
	db sqlite.Database
}

// NewController creates a new PowerSync controller.
func NewController(db sqlite.Database) *Controller {
	return &Controller{db: db}
}

// Control sends a control command and returns the resulting instructions.
// The powersync_control function always expects 2 arguments (op, payload).
func (c *Controller) Control(op ControlOp, payload any) ([]Instruction, error) {
	// Determine how to pass the payload to SQLite
	var sqlPayload any
	if payload == nil {
		sqlPayload = nil // SQL NULL
	} else {
		switch p := payload.(type) {
		case string:
			sqlPayload = p
		case []byte:
			sqlPayload = p
		default:
			// Marshal structs/maps to JSON string
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("marshal payload: %w", err)
			}
			sqlPayload = string(jsonBytes)
		}
	}

	var result []byte
	err := c.db.QueryRow("SELECT powersync_control(?, ?)", string(op), sqlPayload).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("powersync_control(%s): %w", op, err)
	}

	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var instructions []Instruction
	if err := json.Unmarshal(result, &instructions); err != nil {
		// Try single instruction
		var single Instruction
		if err := json.Unmarshal(result, &single); err != nil {
			return nil, fmt.Errorf("unmarshal instructions: %w", err)
		}
		instructions = []Instruction{single}
	}

	return instructions, nil
}

// Start begins a sync stream with the given parameters.
func (c *Controller) Start(req StartRequest) ([]Instruction, error) {
	return c.Control(OpStart, req)
}

// Stop stops the current sync stream.
func (c *Controller) Stop() ([]Instruction, error) {
	return c.Control(OpStop, nil)
}

// SendTextLine forwards a JSON line from the sync service.
func (c *Controller) SendTextLine(line string) ([]Instruction, error) {
	return c.Control(OpLineText, line)
}

// SendBinaryLine forwards a BSON line from the sync service.
func (c *Controller) SendBinaryLine(data []byte) ([]Instruction, error) {
	return c.Control(OpLineBinary, data)
}

// NotifyConnection notifies of a connection state change.
func (c *Controller) NotifyConnection(event ConnectionEvent) ([]Instruction, error) {
	return c.Control("connection", string(event))
}

// NotifyTokenRefreshed notifies that the auth token was refreshed.
func (c *Controller) NotifyTokenRefreshed() ([]Instruction, error) {
	return c.Control(OpRefreshedToken, nil)
}

// NotifyUploadCompleted notifies that CRUD upload completed.
func (c *Controller) NotifyUploadCompleted() ([]Instruction, error) {
	return c.Control(OpCompletedUpload, nil)
}
