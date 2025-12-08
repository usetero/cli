package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// isTokenExpired checks if a JWT access token is expired.
// Returns true if expired or if the token can't be parsed.
func isTokenExpired(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true
	}

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return true
	}

	// Add 30 second buffer to avoid edge cases
	return time.Now().Unix() > claims.Exp-30
}
