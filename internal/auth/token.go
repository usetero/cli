package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenClaims holds the parsed claims from a JWT access token.
type TokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	OrgID string `json:"org_id"`
	Exp   int64  `json:"exp"`
}

// IsExpired returns true if the token has expired (with a 30-second buffer).
func (c TokenClaims) IsExpired() bool {
	return time.Now().Unix() > c.Exp-30
}

// ExpiresAt returns the token's expiration time.
func (c TokenClaims) ExpiresAt() time.Time {
	return time.Unix(c.Exp, 0)
}

// ParseToken extracts claims from a JWT access token without signature verification.
func ParseToken(token string) (TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return TokenClaims{}, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return TokenClaims{}, fmt.Errorf("failed to decode token: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return TokenClaims{}, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}
