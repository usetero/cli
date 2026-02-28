package chat

import (
	"bufio"
	"io"
	"strings"
)

const streamDone = "[DONE]"

const (
	streamScannerInitialBuffer = 64 * 1024
	streamScannerMaxBuffer     = 4 * 1024 * 1024
)

type streamDataHandler func(data []byte, done bool) error

func readStream(r io.Reader, handler streamDataHandler) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, streamScannerInitialBuffer), streamScannerMaxBuffer)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == streamDone {
			if err := handler(nil, true); err != nil {
				return err
			}
			continue
		}

		if err := handler([]byte(data), false); err != nil {
			return err
		}
	}

	return scanner.Err()
}
