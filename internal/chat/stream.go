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

// readStream reads SSE events from the reader and calls the handler for each.
func readStream(r io.Reader, handler Handler) error {
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
			return handler(Event{Done: true})
		}

		// Parse event
		var event Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("parse event: %w", err)
		}

		if err := handler(event); err != nil {
			return err
		}
	}

	return scanner.Err()
}
