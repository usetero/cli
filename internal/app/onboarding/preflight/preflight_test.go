package preflight

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func TestResolveOrg(t *testing.T) {
	t.Parallel()

	orgs := []domain.Organization{
		{ID: "org-1", Name: "One"},
		{ID: "org-2", Name: "Two"},
	}

	got := resolveOrg(orgs, "org-2")
	if got == nil || got.ID != "org-2" {
		t.Fatalf("expected org-2, got %#v", got)
	}

	got = resolveOrg(orgs, "")
	if got != nil {
		t.Fatalf("expected nil when multiple orgs and no preference")
	}

	got = resolveOrg(orgs[:1], "")
	if got == nil || got.ID != "org-1" {
		t.Fatalf("expected single org fallback, got %#v", got)
	}
}

func TestResolveAccount(t *testing.T) {
	t.Parallel()

	accounts := []domain.Account{
		{ID: "acc-1", Name: "One"},
		{ID: "acc-2", Name: "Two"},
	}

	got := resolveAccount(accounts, "acc-2")
	if got == nil || got.ID != "acc-2" {
		t.Fatalf("expected acc-2, got %#v", got)
	}

	got = resolveAccount(accounts, "")
	if got != nil {
		t.Fatalf("expected nil when multiple accounts and no preference")
	}

	got = resolveAccount(accounts[:1], "")
	if got == nil || got.ID != "acc-1" {
		t.Fatalf("expected single account fallback, got %#v", got)
	}
}

func TestPreflightOutcomeForError(t *testing.T) {
	t.Parallel()

	outcome, _ := preflightOutcomeForError(context.DeadlineExceeded)
	if outcome != bootstrap.PreflightOutcomeFailed {
		t.Fatalf("expected failed outcome for deadline exceeded, got %v", outcome)
	}

	outcome, _ = preflightOutcomeForError(errors.New("boom"))
	if outcome != bootstrap.PreflightOutcomeInconclusive {
		t.Fatalf("expected inconclusive outcome for generic error, got %v", outcome)
	}
}

func TestUserFromAccessToken(t *testing.T) {
	t.Parallel()

	token := testJWT(map[string]any{
		"sub":   "user-123",
		"email": "user@example.com",
		"exp":   int64(4102444800), // 2100-01-01
	})
	user := userFromAccessToken(token)
	if user == nil {
		t.Fatal("expected user from token")
	}
	if user.ID != "user-123" {
		t.Fatalf("user.ID = %q, want %q", user.ID, "user-123")
	}
	if user.Email != "user@example.com" {
		t.Fatalf("user.Email = %q, want %q", user.Email, "user@example.com")
	}
}

func TestUserFromAccessTokenMissingSubReturnsNil(t *testing.T) {
	t.Parallel()

	token := testJWT(map[string]any{
		"email": "user@example.com",
		"exp":   int64(4102444800),
	})
	if got := userFromAccessToken(token); got != nil {
		t.Fatalf("expected nil user for missing sub, got %#v", got)
	}
}

func TestUserFromAccessTokenInvalidTokenReturnsNil(t *testing.T) {
	t.Parallel()

	if got := userFromAccessToken("not-a-jwt"); got != nil {
		t.Fatalf("expected nil user for invalid token, got %#v", got)
	}
}

func testJWT(claims map[string]any) string {
	headerJSON, _ := json.Marshal(map[string]string{"alg": "none", "typ": "JWT"})
	payloadJSON, _ := json.Marshal(claims)
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	return header + "." + payload + ".sig"
}
