package powersync

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"

	"github.com/usetero/cli/internal/log"
)

const defaultCaptureMaxBytes int64 = 25 * 1024 * 1024

// StreamCapture receives raw NDJSON sync-stream lines.
//
// Implementations must be best-effort and never panic.
type StreamCapture interface {
	CaptureLine(line []byte)
	Close() error
}

type ndjsonStreamCapture struct {
	path     string
	maxBytes int64
	scope    log.Scope

	mu       sync.Mutex
	file     *os.File
	writer   *bufio.Writer
	written  int64
	disabled bool
}

var _ StreamCapture = (*ndjsonStreamCapture)(nil)

// NewNDJSONStreamCapture creates a best-effort raw line capture sink.
//
// The capture file is created with permissions 0600, and parent directories
// are created with permissions 0700.
func NewNDJSONStreamCapture(path string, maxBytes int64, scope log.Scope) (StreamCapture, error) {
	if maxBytes <= 0 {
		maxBytes = defaultCaptureMaxBytes
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	return &ndjsonStreamCapture{
		path:     path,
		maxBytes: maxBytes,
		scope:    scope.Child("streamcapture"),
		file:     f,
		writer:   bufio.NewWriter(f),
		written:  info.Size(),
	}, nil
}

func (c *ndjsonStreamCapture) CaptureLine(line []byte) {
	if len(line) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.disabled || c.writer == nil {
		return
	}

	writeLen := int64(len(line) + 1)
	if c.written+writeLen > c.maxBytes {
		c.scope.Warn("disabled stream capture after reaching size limit", "path", c.path, "max_bytes", c.maxBytes)
		c.disabled = true
		return
	}

	if _, err := c.writer.Write(line); err != nil {
		c.scope.Error("disabled stream capture after write failure", "path", c.path, "error", err)
		c.disabled = true
		return
	}
	if err := c.writer.WriteByte('\n'); err != nil {
		c.scope.Error("disabled stream capture after write failure", "path", c.path, "error", err)
		c.disabled = true
		return
	}
	if err := c.writer.Flush(); err != nil {
		c.scope.Error("disabled stream capture after flush failure", "path", c.path, "error", err)
		c.disabled = true
		return
	}

	c.written += writeLen
}

func (c *ndjsonStreamCapture) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.writer == nil {
		return nil
	}

	if err := c.writer.Flush(); err != nil {
		_ = c.file.Close()
		c.writer = nil
		c.file = nil
		return err
	}
	err := c.file.Close()
	c.writer = nil
	c.file = nil
	return err
}
