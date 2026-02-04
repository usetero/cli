package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// streamDone is the SSE sentinel value indicating the stream is complete.
const streamDone = "[DONE]"

// eventHandler is called for each event in the response stream.
type eventHandler func(event) error

// readStream reads SSE events from the reader and calls the handler for each.
func readStream(r io.Reader, handler eventHandler) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data line
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Stream complete
		if data == streamDone {
			return handler(event{Done: true})
		}

		// Parse event
		var e event
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			return fmt.Errorf("parse event: %w", err)
		}

		// Reject empty/unknown event types
		if e.Type == "" && !e.Done {
			return fmt.Errorf("received empty event type, raw data: %s", data)
		}

		if err := handler(e); err != nil {
			return err
		}
	}

	return scanner.Err()
}
