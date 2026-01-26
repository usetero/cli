//go:build correctness

package workos_test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/keyring"
)

// Correctness tests verify JWT claims from real WorkOS device auth tokens.
// Run on-demand to verify WorkOS configuration and token behavior.
//
// Prerequisites:
//   1. Login: task auth:login
//   2. Run:   task test:correctness
//
// Check status: task auth:status

func TestCorrectness_DeviceAuth_JWT_Claims(t *testing.T) {
	namespace := os.Getenv("TERO_NAMESPACE")
	if namespace == "" {
		namespace = "api.usetero.dev" // dev default
	}

	// Get JWT from keychain (stored by `task auth:login`)
	storage := keyring.New(namespace)
	token, err := storage.Get("access_token")
	if err != nil || token == "" {
		t.Skip("Not logged in. Run: task auth:login")
	}

	t.Run("JWT has valid structure", func(t *testing.T) {
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("Invalid JWT structure: expected 3 parts, got %d", len(parts))
		}
		t.Logf("JWT has valid 3-part structure")
	})

	t.Run("JWT payload is parseable", func(t *testing.T) {
		claims := parseJWTClaims(t, token)
		t.Logf("JWT claims: %+v", claims)
	})

	t.Run("JWT has expected standard claims", func(t *testing.T) {
		claims := parseJWTClaims(t, token)

		// Check issuer
		if iss, ok := claims["iss"].(string); ok {
			t.Logf("Issuer (iss): %s", iss)
			if !strings.Contains(iss, "workos") {
				t.Errorf("Expected WorkOS issuer, got: %s", iss)
			}
		} else {
			t.Error("Missing 'iss' claim")
		}

		// Check subject (user ID)
		if sub, ok := claims["sub"].(string); ok {
			t.Logf("Subject (sub): %s", sub)
		} else {
			t.Error("Missing 'sub' claim")
		}

		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			t.Logf("Expires (exp): %v", exp)
		} else {
			t.Error("Missing 'exp' claim")
		}
	})

	t.Run("JWT has audience claim", func(t *testing.T) {
		claims := parseJWTClaims(t, token)

		aud, hasAud := claims["aud"]
		if !hasAud {
			t.Fatal("Missing 'aud' claim - WorkOS JWT Template needs to be configured")
		}

		t.Logf("Audience (aud): %v (type: %T)", aud, aud)

		// Audience can be a string or array of strings
		switch v := aud.(type) {
		case string:
			t.Logf("Single audience: %s", v)
		case []interface{}:
			for i, a := range v {
				t.Logf("Audience[%d]: %v", i, a)
			}
		default:
			t.Errorf("Unexpected audience type: %T", aud)
		}
	})

	t.Run("JWT audience includes expected endpoints", func(t *testing.T) {
		claims := parseJWTClaims(t, token)

		aud, hasAud := claims["aud"]
		if !hasAud {
			t.Skip("No 'aud' claim - configure WorkOS JWT Template first")
		}

		// Get expected audiences from environment or use defaults
		expectedAPI := os.Getenv("TERO_API_URL")
		if expectedAPI == "" {
			expectedAPI = "https://api.usetero.dev"
		}
		expectedPowerSync := os.Getenv("POWERSYNC_URL")
		if expectedPowerSync == "" {
			expectedPowerSync = "https://powersync.usetero.dev"
		}

		audiences := audienceToSlice(aud)
		t.Logf("Checking for audiences: API=%s, PowerSync=%s", expectedAPI, expectedPowerSync)
		t.Logf("Token audiences: %v", audiences)

		hasAPI := containsAudience(audiences, expectedAPI)
		hasPowerSync := containsAudience(audiences, expectedPowerSync)

		if !hasAPI {
			t.Errorf("JWT missing Tero API audience: %s", expectedAPI)
		}
		if !hasPowerSync {
			t.Errorf("JWT missing PowerSync audience: %s", expectedPowerSync)
		}

		if hasAPI && hasPowerSync {
			t.Log("JWT has both required audiences")
		}
	})

	t.Run("JWT has organization claim if scoped", func(t *testing.T) {
		claims := parseJWTClaims(t, token)

		if orgID, ok := claims["org_id"].(string); ok {
			t.Logf("Organization ID (org_id): %s", orgID)
		} else {
			t.Log("No 'org_id' claim - token is user-scoped (not org-scoped)")
		}
	})
}

// parseJWTClaims extracts claims from a JWT without verifying signature.
// This is safe for testing since we just want to inspect the claims.
func parseJWTClaims(t *testing.T, token string) map[string]interface{} {
	t.Helper()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid JWT: expected 3 parts, got %d", len(parts))
	}

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWT payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("Failed to parse JWT claims: %v", err)
	}

	return claims
}

// audienceToSlice converts the aud claim (string or []string) to a slice.
func audienceToSlice(aud interface{}) []string {
	switch v := aud.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, a := range v {
			if s, ok := a.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}

// containsAudience checks if the audience list contains the expected value.
func containsAudience(audiences []string, expected string) bool {
	for _, aud := range audiences {
		if aud == expected {
			return true
		}
	}
	return false
}
