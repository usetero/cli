package extension

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
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

// Controller wraps the powersync_control SQLite function with type safety.
// It holds a dedicated database connection to ensure all operations use the
// same connection, since the PowerSync extension maintains per-connection state.
type Controller struct {
	db   sqlite.DB
	conn *sql.Conn // dedicated connection for state consistency
}

// NewController creates a new PowerSync controller.
// Call Close() when done to release the dedicated connection.
func NewController(db sqlite.DB) *Controller {
	return &Controller{db: db}
}

// Close releases the dedicated connection.
func (c *Controller) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Control sends a control command and returns the resulting instructions.
// The powersync_control function always expects 2 arguments (op, payload).
// All operations use a dedicated connection to ensure state consistency.
func (c *Controller) Control(ctx context.Context, op ControlOp, payload any) ([]Instruction, error) {
	// Lazily acquire a dedicated connection on first use.
	// This connection is held for the lifetime of the Controller to ensure
	// all powersync_control calls see the same extension state.
	if c.conn == nil {
		conn, err := c.db.Raw().Conn(ctx)
		if err != nil {
			return nil, fmt.Errorf("acquire connection: %w", err)
		}
		c.conn = conn
	}

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
	err := c.conn.QueryRowContext(ctx, "SELECT powersync_control(?, ?)", string(op), sqlPayload).Scan(&result)
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
func (c *Controller) Start(ctx context.Context, req StartRequest) ([]Instruction, error) {
	return c.Control(ctx, OpStart, req)
}

// Stop stops the current sync stream.
func (c *Controller) Stop(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpStop, nil)
}

// SendTextLine forwards a JSON line from the sync service.
func (c *Controller) SendTextLine(ctx context.Context, line string) ([]Instruction, error) {
	return c.Control(ctx, OpLineText, line)
}

// SendBinaryLine forwards a BSON line from the sync service.
func (c *Controller) SendBinaryLine(ctx context.Context, data []byte) ([]Instruction, error) {
	return c.Control(ctx, OpLineBinary, data)
}

// NotifyConnection notifies of a connection state change.
func (c *Controller) NotifyConnection(ctx context.Context, event ConnectionEvent) ([]Instruction, error) {
	return c.Control(ctx, "connection", string(event))
}

// NotifyTokenRefreshed notifies that the auth token was refreshed.
func (c *Controller) NotifyTokenRefreshed(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpRefreshedToken, nil)
}

// NotifyUploadCompleted notifies that CRUD upload completed.
func (c *Controller) NotifyUploadCompleted(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpCompletedUpload, nil)
}
