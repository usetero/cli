package chat

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// StreamErrorClass is a normalized category for stream failures.
type StreamErrorClass string

const (
	StreamErrorClassCancelled StreamErrorClass = "cancelled"
	StreamErrorClassTimeout   StreamErrorClass = "timeout"
	StreamErrorClassProtocol  StreamErrorClass = "protocol_error"
	StreamErrorClassRequest   StreamErrorClass = "request_error"
	StreamErrorClassServer    StreamErrorClass = "server_error"
	StreamErrorClassUnknown   StreamErrorClass = "unknown"
)

var httpStatusCodePattern = regexp.MustCompile(`\b([1-5][0-9]{2})\b`)

// ClassifyStreamError maps raw stream errors into stable operational buckets.
func ClassifyStreamError(err error) StreamErrorClass {
	if err == nil {
		return StreamErrorClassUnknown
	}
	if errors.Is(err, context.Canceled) {
		return StreamErrorClassCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return StreamErrorClassTimeout
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "user_cancelled"),
		strings.Contains(msg, "context canceled"),
		strings.Contains(msg, "canceled"):
		return StreamErrorClassCancelled
	case strings.Contains(msg, "deadline exceeded"),
		strings.Contains(msg, "timeout"):
		return StreamErrorClassTimeout
	case strings.Contains(msg, "protocol error"),
		strings.Contains(msg, "parse event"):
		return StreamErrorClassProtocol
	case strings.Contains(msg, "server error:"),
		strings.Contains(msg, "chat api error"):
		if status, ok := firstHTTPStatusCode(msg); ok {
			switch {
			case status >= 400 && status <= 499:
				return StreamErrorClassRequest
			case status >= 500 && status <= 599:
				return StreamErrorClassServer
			}
		}
		return StreamErrorClassServer
	default:
		return StreamErrorClassUnknown
	}
}

func firstHTTPStatusCode(msg string) (int, bool) {
	matches := httpStatusCodePattern.FindStringSubmatch(msg)
	if len(matches) < 2 {
		return 0, false
	}
	status, err := strconv.Atoi(matches[1])
	if err != nil || status < 100 || status > 599 {
		return 0, false
	}
	return status, true
}

// UserFacingStreamError returns concise error copy suitable for toast UI.
func UserFacingStreamError(err error) string {
	switch ClassifyStreamError(err) {
	case StreamErrorClassCancelled:
		return "Request was cancelled."
	case StreamErrorClassTimeout:
		return "The response timed out. Please try again."
	case StreamErrorClassProtocol:
		return "The chat service returned an unexpected stream format. Please retry."
	case StreamErrorClassRequest:
		return "The request was rejected by the chat service. Please retry."
	case StreamErrorClassServer:
		return "The chat service returned an internal error. Please try again."
	default:
		return "Something went wrong while streaming the response. Please try again."
	}
}
