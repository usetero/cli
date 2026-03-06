package main

import (
	"encoding/base64"
	"testing"
)

func TestUserIDFromAccessToken(t *testing.T) {
	t.Parallel()

	token := "x." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user_123"}`)) + ".y"
	userID, err := userIDFromAccessToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user_123" {
		t.Fatalf("unexpected user id: %s", userID)
	}
}

func TestUserIDFromAccessToken_InvalidToken(t *testing.T) {
	t.Parallel()

	_, err := userIDFromAccessToken("not-a-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserIDFromAccessToken_EmptySubject(t *testing.T) {
	t.Parallel()

	token := "x." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":""}`)) + ".y"
	_, err := userIDFromAccessToken(token)
	if err == nil {
		t.Fatal("expected error")
	}
}
