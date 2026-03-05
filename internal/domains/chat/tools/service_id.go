package tools

import (
	"fmt"
	"strings"
)

type ServiceID string

func ParseServiceID(raw string) (ServiceID, error) {
	id := ServiceID(strings.TrimSpace(raw))
	if id == "" {
		return "", fmt.Errorf("service_id is required")
	}
	return id, nil
}
