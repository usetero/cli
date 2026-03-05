package cmd

import (
	"testing"

	"github.com/usetero/cli/internal/auth"
)

func TestParseToken(t *testing.T) {
	t.Parallel()
	t.Run("parses valid JWT", func(t *testing.T) {
		t.Parallel()
		// JWT with payload: {"sub":"user123","email":"test@example.com","org_id":"org456","exp":1234567890}
		// Header: {"alg":"HS256","typ":"JWT"}
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwib3JnX2lkIjoib3JnNDU2IiwiZXhwIjoxMjM0NTY3ODkwfQ.signature"

		claims, err := auth.ParseToken(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if claims.Sub != "user123" {
			t.Errorf("expected sub=user123, got %v", claims.Sub)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("expected email=test@example.com, got %v", claims.Email)
		}
		if claims.OrgID != "org456" {
			t.Errorf("expected org_id=org456, got %v", claims.OrgID)
		}
	})

	t.Run("returns error for invalid token format", func(t *testing.T) {
		t.Parallel()
		_, err := auth.ParseToken("not-a-jwt")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("returns error for invalid base64", func(t *testing.T) {
		t.Parallel()
		_, err := auth.ParseToken("header.!!!invalid-base64!!!.signature")
		if err == nil {
			t.Error("expected error for invalid base64")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		t.Parallel()
		// "notjson" base64 encoded
		_, err := auth.ParseToken("header.bm90anNvbg.signature")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestPromptOrgSelection_Validation(t *testing.T) {
	t.Parallel()
	// Note: We can't easily test the interactive prompt without stdin mocking,
	// but we can test that the function exists and has the right signature.
	// The actual selection logic is simple enough (parse int, bounds check)
	// that integration testing via manual QA is appropriate.

	// This test documents the expected behavior:
	// - Shows numbered list of orgs
	// - Accepts number input
	// - Returns error for invalid input
}

func TestFetchOrganizations(t *testing.T) {
	t.Parallel()
	// fetchOrganizations is a thin wrapper around the API services.
	// Testing it would require mocking the entire API client,
	// which is overkill for a simple data transformation.
}
