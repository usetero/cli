package extension

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// ControlOp represents an operation for powersync_control.
type ControlOp string

const (
	OpStart           ControlOp = "start"
	OpStop            ControlOp = "stop"
	OpLineText        ControlOp = "line_text"
	OpLineBinary      ControlOp = "line_binary"
	OpRefreshedToken  ControlOp = "refreshed_token"
	OpCompletedUpload ControlOp = "completed_upload"
)

// ConnectionEvent represents a connection state change.
type ConnectionEvent string

const (
	ConnectionEstablished ConnectionEvent = "established"
	ConnectionEnded       ConnectionEvent = "end"
)

// StartRequest is the payload for start.
type StartRequest struct {
	Parameters      map[string]any  `json:"parameters,omitempty"`
	Schema          json.RawMessage `json:"schema,omitempty"`
	IncludeDefaults bool            `json:"include_defaults"`
	ActiveStreams   []StreamKey     `json:"active_streams,omitempty"`
}

// StreamKey identifies a stream subscription.
type StreamKey struct {
	Name       string `json:"name"`
	Parameters string `json:"parameters,omitempty"`
}

// Controller wraps powersync_control with typed operations.
// It keeps a dedicated SQL connection for extension state consistency.
type Controller struct {
	db   *sqlite.DB
	conn *sql.Conn
}

func NewController(db *sqlite.DB) *Controller {
	return &Controller{db: db}
}

func (c *Controller) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Control sends a control command and decodes resulting instructions.
func (c *Controller) Control(ctx context.Context, op ControlOp, payload any) ([]Instruction, error) {
	if c.conn == nil {
		conn, err := c.db.Raw().Conn(ctx)
		if err != nil {
			return nil, fmt.Errorf("acquire connection: %w", err)
		}
		c.conn = conn
	}

	var sqlPayload any
	if payload == nil {
		sqlPayload = nil
	} else {
		switch p := payload.(type) {
		case string, []byte:
			sqlPayload = p
		default:
			j, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("marshal payload: %w", err)
			}
			sqlPayload = string(j)
		}
	}

	var result []byte
	err := c.conn.QueryRowContext(ctx, "SELECT powersync_control(?, ?)", string(op), sqlPayload).Scan(&result)
	if err != nil {
		// The extension reports this state via error string only.
		if strings.Contains(err.Error(), "No iteration is active") {
			return nil, fmt.Errorf("%w: %v", ErrNoActiveIteration, err)
		}
		if errors.Is(err, sql.ErrConnDone) {
			return nil, err
		}
		return nil, fmt.Errorf("powersync_control(%s): %w", op, err)
	}
	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var instructions []Instruction
	if err := json.Unmarshal(result, &instructions); err != nil {
		var single Instruction
		if err2 := json.Unmarshal(result, &single); err2 != nil {
			return nil, fmt.Errorf("unmarshal instructions: %w", err)
		}
		instructions = []Instruction{single}
	}
	return instructions, nil
}

func (c *Controller) Start(ctx context.Context, req StartRequest) ([]Instruction, error) {
	return c.Control(ctx, OpStart, req)
}

func (c *Controller) Stop(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpStop, nil)
}

func (c *Controller) SendTextLine(ctx context.Context, line string) ([]Instruction, error) {
	return c.Control(ctx, OpLineText, line)
}

func (c *Controller) NotifyConnection(ctx context.Context, event ConnectionEvent) ([]Instruction, error) {
	return c.Control(ctx, "connection", string(event))
}

func (c *Controller) NotifyTokenRefreshed(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpRefreshedToken, nil)
}

func (c *Controller) NotifyUploadCompleted(ctx context.Context) ([]Instruction, error) {
	return c.Control(ctx, OpCompletedUpload, nil)
}
