package tools

import (
	"fmt"
	"strings"
)

type PolicyID string

func ParsePolicyID(raw string) (PolicyID, error) {
	id := PolicyID(strings.TrimSpace(raw))
	if id == "" {
		return "", fmt.Errorf("policy_id is required")
	}
	return id, nil
}
